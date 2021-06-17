package data

import (
	"context"
	"net/http"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/sync/errgroup"

	pb_engine "github.com/beneath-hq/beneath/infra/engine/proto"
	"github.com/beneath-hq/beneath/models"
	"github.com/beneath-hq/beneath/pkg/jsonutil"
	"github.com/beneath-hq/beneath/pkg/timeutil"
	pb "github.com/beneath-hq/beneath/server/data/grpc/proto"
	"github.com/beneath-hq/beneath/services/middleware"
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
	WriteID   uuid.UUID    `json:"write_id,omitempty"`
	Instances []writeUsage `json:"instances,omitempty"`
}

type writeUsage struct {
	InstanceID   uuid.UUID `json:"instance,omitempty"`
	RecordsCount int       `json:"records,omitempty"`
	BytesWritten int       `json:"bytes,omitempty"`
}

// HandleWrite handles a write request
func (s *Service) HandleWrite(ctx context.Context, req *WriteRequest) (*WriteResponse, *Error) {
	// get auth
	secret := middleware.GetSecret(ctx)
	if secret == nil || secret.IsAnonymous() {
		return nil, newError(http.StatusForbidden, "not authenticated")
	}

	// check quota
	err := s.Usage.CheckWriteQuota(ctx, secret)
	if err != nil {
		return nil, newError(http.StatusTooManyRequests, err.Error())
	}

	// check number of instances
	if len(req.InstanceRecords) > maxInstancesPerWrite {
		return nil, newErrorf(http.StatusBadRequest, "a single request cannot write to more than %d instances", maxInstancesPerWrite)
	}

	// get instances concurrently
	mu := &sync.Mutex{}
	instances := make(map[uuid.UUID]*models.CachedInstance, len(req.InstanceRecords))
	group, cctx := errgroup.WithContext(ctx)
	for instanceID, records := range req.InstanceRecords {
		instanceID := instanceID // see https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		records := records
		group.Go(func() error {
			// check records supplied
			if records.Len() == 0 {
				return newErrorf(http.StatusBadRequest, "records are empty for instance_id=%s", instanceID.String())
			}

			// get table info
			table := s.Tables.FindCachedInstance(cctx, instanceID)
			if table == nil {
				return newErrorf(http.StatusNotFound, "table not found for instance_id=%s", instanceID.String())
			}

			// check not already a committed batch table
			if table.Final {
				return newErrorf(http.StatusBadRequest, "instance_id=%s closed for further writes because it has been marked final", instanceID.String())
			}

			// check permissions
			perms := s.Permissions.TablePermissionsForSecret(cctx, secret, table.TableID, table.ProjectID, table.Public)
			if !perms.Write {
				return newErrorf(http.StatusForbidden, "secret doesn't grant right to write to table_id=%s", table.TableID.String())
			}

			// set table
			mu.Lock()
			instances[instanceID] = table
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
	instanceUsage := make([]writeUsage, 0, len(req.InstanceRecords))
	for instanceID, records := range req.InstanceRecords {
		// check the batch length is valid
		err = s.Engine.CheckBatchLength(records.Len())
		if err != nil {
			return nil, newError(http.StatusBadRequest, err.Error())
		}

		// check records
		table := instances[instanceID]
		pbs, bytes, err := records.prepare(s, instanceID, table)
		if err != nil {
			return nil, err
		}

		// add
		instanceRecords = append(instanceRecords, &pb.InstanceRecords{
			InstanceId: instanceID.Bytes(),
			Records:    pbs,
		})
		instanceUsage = append(instanceUsage, writeUsage{
			InstanceID:   instanceID,
			BytesWritten: bytes,
			RecordsCount: records.Len(),
		})
	}

	// write request to engine
	writeID := uuid.NewV4()
	err = s.QueueWriteRequest(ctx, &pb_engine.WriteRequest{
		WriteId:         writeID.Bytes(),
		InstanceRecords: instanceRecords,
	})
	if err != nil {
		return nil, newError(http.StatusBadRequest, err.Error())
	}

	// track write usage
	for _, im := range instanceUsage {
		table := instances[im.InstanceID]
		s.Usage.TrackWrite(ctx, secret, table.TableID, im.InstanceID, int64(im.RecordsCount), int64(im.BytesWritten))
	}

	// set log payload
	payload := writeTags{
		WriteID:   writeID,
		Instances: instanceUsage,
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
func (r *WriteRecords) prepare(s *Service, instanceID uuid.UUID, table *models.CachedInstance) ([]*pb.Record, int, *Error) {
	if r.PB != nil {
		return r.preparePB(s, instanceID, table)
	}
	return r.prepareJSON(s, instanceID, table)
}

func (r *WriteRecords) preparePB(s *Service, instanceID uuid.UUID, table *models.CachedInstance) ([]*pb.Record, int, *Error) {
	// validate each record
	bytes := 0
	for i, record := range r.PB {
		// set timestamp to current timestamp if it's 0
		if record.Timestamp == 0 {
			record.Timestamp = timeutil.UnixMilli(time.Now())
		}

		// check it decodes
		structured, err := table.Codec.UnmarshalAvro(record.AvroData)
		if err != nil {
			return nil, 0, newErrorf(http.StatusBadRequest, "error for instance_id=%s, record at index %d: %v", instanceID.String(), i, err.Error())
		}

		// check record size
		err = s.Engine.CheckRecordSize(table, structured, len(record.AvroData))
		if err != nil {
			return nil, 0, newErrorf(http.StatusBadRequest, "error for instance_id=%s, record at index %d: %v", instanceID.String(), i, err.Error())
		}

		// increment bytes written
		bytes += len(record.AvroData)
	}

	return r.PB, bytes, nil
}

func (r *WriteRecords) prepareJSON(s *Service, instanceID uuid.UUID, table *models.CachedInstance) ([]*pb.Record, int, *Error) {
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
		native, err := table.Codec.ConvertFromJSONTypes(obj)
		if err != nil {
			return nil, 0, newErrorf(http.StatusBadRequest, "error encoding record at index %d for instance %s: %s", idx, instanceID.String(), err.Error())
		}

		// encode as avro
		avroData, err := table.Codec.MarshalAvro(native)
		if err != nil {
			return nil, 0, newErrorf(http.StatusBadRequest, "error encoding record at index %d for instance %s: %s", idx, instanceID.String(), err.Error())
		}

		err = s.Engine.CheckRecordSize(table, native, len(avroData))
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
