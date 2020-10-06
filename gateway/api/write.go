package api

import (
	"context"
	"net/http"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/sync/errgroup"

	"gitlab.com/beneath-hq/beneath/control/entity"
	pb_engine "gitlab.com/beneath-hq/beneath/engine/proto"
	pb "gitlab.com/beneath-hq/beneath/gateway/grpc/proto"
	"gitlab.com/beneath-hq/beneath/gateway/util"
	"gitlab.com/beneath-hq/beneath/hub"
	"gitlab.com/beneath-hq/beneath/internal/middleware"
	"gitlab.com/beneath-hq/beneath/pkg/jsonutil"
	"gitlab.com/beneath-hq/beneath/pkg/timeutil"
)

// WriteRequest is a request to HandleWrite
type WriteRequest struct {
	InstanceRecords map[uuid.UUID]*WriteRecords
}

// WriteResponse is a result from HandleWrite
type WriteResponse struct {
	WriteID uuid.UUID
}

const (
	maxInstancesPerWrite = 5
)

type writeTags struct {
	WriteID   uuid.UUID      `json:"write_id,omitempty"`
	Instances []writeMetrics `json:"instances,omitempty"`
}

type writeMetrics struct {
	InstanceID   uuid.UUID `json:"instance,omitempty"`
	RecordsCount int       `json:"records,omitempty"`
	BytesWritten int       `json:"bytes,omitempty"`
}

// HandleWrite handles a write request
func HandleWrite(ctx context.Context, req *WriteRequest) (*WriteResponse, *Error) {
	// get auth
	secret := middleware.GetSecret(ctx)
	if secret == nil {
		return nil, newError(http.StatusForbidden, "not authenticated")
	}

	// check quota
	err := util.CheckWriteQuota(ctx, secret)
	if err != nil {
		return nil, newError(http.StatusTooManyRequests, err.Error())
	}

	// check number of instances
	if len(req.InstanceRecords) > maxInstancesPerWrite {
		return nil, newErrorf(http.StatusBadRequest, "a single request cannot write to more than %d instances", maxInstancesPerWrite)
	}

	// get instances concurrently
	mu := &sync.Mutex{}
	instances := make(map[uuid.UUID]*entity.CachedStream, len(req.InstanceRecords))
	group, cctx := errgroup.WithContext(ctx)
	for instanceID, records := range req.InstanceRecords {
		instanceID := instanceID // see https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		records := records
		group.Go(func() error {
			// check records supplied
			if records.Len() == 0 {
				return newErrorf(http.StatusBadRequest, "records are empty for instance_id=%s", instanceID.String())
			}

			// get stream info
			stream := entity.FindCachedStreamByCurrentInstanceID(cctx, instanceID)
			if stream == nil {
				return newErrorf(http.StatusNotFound, "stream not found for instance_id=%s", instanceID.String())
			}

			// check not already a committed batch stream
			if stream.Final {
				return newErrorf(http.StatusBadRequest, "instance_id=%s closed for further writes because it has been marked final", instanceID.String())
			}

			// check permissions
			perms := secret.StreamPermissions(cctx, stream.StreamID, stream.ProjectID, stream.Public)
			if !perms.Write {
				return newErrorf(http.StatusForbidden, "secret doesn't grant right to write to stream_id=%s", stream.StreamID.String())
			}

			// set stream
			mu.Lock()
			instances[instanceID] = stream
			mu.Unlock()

			return nil
		})
	}

	// run wait group
	err = group.Wait()
	if err != nil {
		return nil, err.(*Error)
	}

	// process each InstanceRecords
	instanceRecords := make([]*pb.InstanceRecords, 0, len(req.InstanceRecords))
	instanceMetrics := make([]writeMetrics, 0, len(req.InstanceRecords))
	for instanceID, records := range req.InstanceRecords {
		// check the batch length is valid
		err = hub.Engine.CheckBatchLength(records.Len())
		if err != nil {
			return nil, newError(http.StatusBadRequest, err.Error())
		}

		// check records
		stream := instances[instanceID]
		pbs, bytes, err := records.prepare(instanceID, stream)
		if err != nil {
			return nil, err
		}

		// add
		instanceRecords = append(instanceRecords, &pb.InstanceRecords{
			InstanceId: instanceID.Bytes(),
			Records:    pbs,
		})
		instanceMetrics = append(instanceMetrics, writeMetrics{
			InstanceID:   instanceID,
			BytesWritten: bytes,
			RecordsCount: records.Len(),
		})
	}

	// write request to engine
	writeID := uuid.NewV4()
	err = hub.Engine.QueueWriteRequest(ctx, &pb_engine.WriteRequest{
		WriteId:         writeID.Bytes(),
		InstanceRecords: instanceRecords,
	})
	if err != nil {
		return nil, newError(http.StatusBadRequest, err.Error())
	}

	// track write metrics
	for _, im := range instanceMetrics {
		stream := instances[im.InstanceID]
		util.TrackWrite(ctx, secret, stream.StreamID, im.InstanceID, int64(im.RecordsCount), int64(im.BytesWritten))
	}

	// set log payload
	payload := writeTags{
		WriteID:   writeID,
		Instances: instanceMetrics,
	}
	middleware.SetTagsPayload(ctx, payload)

	return &WriteResponse{WriteID: writeID}, nil
}

// WriteRecords represents records to write, either as parsed json or protocol buffers
type WriteRecords struct {
	PB   []*pb.Record
	JSON []interface{}
}

// Len returns the number of records wrapped by WriteRecords
func (r *WriteRecords) Len() int {
	if r.PB != nil {
		return len(r.PB)
	}
	return len(r.JSON)
}

// prepares for writing
func (r *WriteRecords) prepare(instanceID uuid.UUID, stream *entity.CachedStream) ([]*pb.Record, int, *Error) {
	if r.PB != nil {
		return r.preparePB(instanceID, stream)
	}
	return r.prepareJSON(instanceID, stream)
}

func (r *WriteRecords) preparePB(instanceID uuid.UUID, stream *entity.CachedStream) ([]*pb.Record, int, *Error) {
	// validate each record
	bytes := 0
	for i, record := range r.PB {
		// set timestamp to current timestamp if it's 0
		if record.Timestamp == 0 {
			record.Timestamp = timeutil.UnixMilli(time.Now())
		}

		// check it decodes
		structured, err := stream.Codec.UnmarshalAvro(record.AvroData)
		if err != nil {
			return nil, 0, newErrorf(http.StatusBadRequest, "error for instance_id=%s, record at index %d: %v", instanceID.String(), i, err.Error())
		}

		// check record size
		err = hub.Engine.CheckRecordSize(stream, structured, len(record.AvroData))
		if err != nil {
			return nil, 0, newErrorf(http.StatusBadRequest, "error for instance_id=%s, record at index %d: %v", instanceID.String(), i, err.Error())
		}

		// increment bytes written
		bytes += len(record.AvroData)
	}

	return r.PB, bytes, nil
}

func (r *WriteRecords) prepareJSON(instanceID uuid.UUID, stream *entity.CachedStream) ([]*pb.Record, int, *Error) {
	// convert objects into records
	records := make([]*pb.Record, len(r.JSON))
	bytesWritten := 0
	for idx, objV := range r.JSON {
		// check it's a map
		obj, ok := objV.(map[string]interface{})
		if !ok {
			return nil, 0, newErrorf(http.StatusBadRequest, "record at index %d for instance %s is not an object", idx, instanceID.String())
		}

		// get timestamp
		timestamp := time.Now()
		meta, ok := obj["@meta"].(map[string]interface{})
		if ok {
			raw := meta["timestamp"]
			if raw != nil {
				raw, err := jsonutil.ParseInt64(meta["timestamp"])
				if err != nil {
					return nil, 0, newErrorf(http.StatusBadRequest, "couldn't parse '@meta.timestamp' as number or numeric string for record at index %d for instance %s", idx, instanceID.String())
				}
				timestamp = timeutil.FromUnixMilli(raw)
			}
		}

		// convert from json types to native types for encoding
		native, err := stream.Codec.ConvertFromJSONTypes(obj)
		if err != nil {
			return nil, 0, newErrorf(http.StatusBadRequest, "error encoding record at index %d for instance %s: %s", idx, instanceID.String(), err.Error())
		}

		// encode as avro
		avroData, err := stream.Codec.MarshalAvro(native)
		if err != nil {
			return nil, 0, newErrorf(http.StatusBadRequest, "error encoding record at index %d for instance %s: %s", idx, instanceID.String(), err.Error())
		}

		err = hub.Engine.CheckRecordSize(stream, native, len(avroData))
		if err != nil {
			return nil, 0, newErrorf(http.StatusBadRequest, "error encoding record at index %d for instance %s: %s", idx, instanceID.String(), err.Error())
		}

		// save the record
		records[idx] = &pb.Record{
			AvroData:  avroData,
			Timestamp: timeutil.UnixMilli(timestamp),
		}

		// increment bytes written
		bytesWritten += len(avroData)
	}

	return records, bytesWritten, nil
}
