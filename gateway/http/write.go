package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	uuid "github.com/satori/go.uuid"

	"github.com/beneath-core/control/entity"
	"github.com/beneath-core/pkg/httputil"
	"github.com/beneath-core/pkg/jsonutil"
	"github.com/beneath-core/internal/middleware"
	"github.com/beneath-core/pkg/timeutil"
	"github.com/beneath-core/internal/hub"
	"github.com/beneath-core/engine"
	pb_engine "github.com/beneath-core/engine/proto"
	"github.com/beneath-core/gateway"
	pb "github.com/beneath-core/gateway/grpc/proto"
)

func postToInstance(w http.ResponseWriter, r *http.Request) error {
	// get auth
	secret := middleware.GetSecret(r.Context())

	// get instance ID
	instanceID, err := uuid.FromString(chi.URLParam(r, "instanceID"))
	if err != nil {
		return httputil.NewError(404, "instance not found -- malformed ID")
	}

	// set log payload
	payload := writeTags{
		InstanceID: instanceID.String(),
	}
	middleware.SetTagsPayload(r.Context(), payload)

	// get stream
	stream := entity.FindCachedStreamByCurrentInstanceID(r.Context(), instanceID)
	if stream == nil {
		return httputil.NewError(404, "stream not found")
	}

	// check not already a committed batch stream
	if stream.Batch && stream.Committed {
		return httputil.NewError(400, "batch has been committed and closed for further writes")
	}

	// check allowed to write stream
	perms := secret.StreamPermissions(r.Context(), stream.StreamID, stream.ProjectID, stream.Public, stream.External)
	if !perms.Write {
		return httputil.NewError(403, "secret doesn't grant right to write this stream")
	}

	// check quota
	usage := gateway.Metrics.GetCurrentUsage(r.Context(), secret.GetOwnerID())
	ok := secret.CheckWriteQuota(usage)
	if !ok {
		return httputil.NewError(429, "you have exhausted your monthly quota")
	}

	// decode json body
	var body interface{}
	err = jsonutil.Unmarshal(r.Body, &body)
	if err != nil {
		return httputil.NewError(400, "request body must be json")
	}

	// get objects passed in body
	var objects []interface{}
	switch bodyT := body.(type) {
	case []interface{}:
		objects = bodyT
	case map[string]interface{}:
		objects = []interface{}{bodyT}
	default:
		return httputil.NewError(400, "request body must be an array or an object")
	}

	// check the batch length is valid
	err = hub.Engine.CheckBatchLength(len(objects))
	if err != nil {
		return httputil.NewError(400, fmt.Sprintf("error encoding batch: %v", err.Error()))
	}

	// convert objects into records
	records := make([]*pb.Record, len(objects))
	bytesWritten := 0
	for idx, objV := range objects {
		// check it's a map
		obj, ok := objV.(map[string]interface{})
		if !ok {
			return httputil.NewError(400, fmt.Sprintf("record at index %d is not an object", idx))
		}

		// get timestamp
		timestamp := time.Now()
		meta, ok := obj["@meta"].(map[string]interface{})
		if ok {
			raw := meta["timestamp"]
			if raw != nil {
				raw, err := jsonutil.ParseInt64(meta["timestamp"])
				if err != nil {
					return httputil.NewError(400, "couldn't parse '@meta.timestamp' as number or numeric string for record at index %d", idx)
				}
				timestamp = timeutil.FromUnixMilli(raw)
			}
		}

		// convert to avro native for encoding
		avroNative, err := stream.Codec.ConvertToAvroNative(obj, true)
		if err != nil {
			return httputil.NewError(400, fmt.Sprintf("error encoding record at index %d: %v", idx, err.Error()))
		}

		// encode as avro
		avroData, err := stream.Codec.MarshalAvro(avroNative)
		if err != nil {
			return httputil.NewError(400, fmt.Sprintf("error encoding record at index %d: %v", idx, err.Error()))
		}

		err = hub.Engine.CheckRecordSize(stream, avroNative, len(avroData))
		if err != nil {
			return httputil.NewError(400, fmt.Sprintf("error encoding record at index %d: %v", idx, err.Error()))
		}

		// save the record
		records[idx] = &pb.Record{
			AvroData:  avroData,
			Timestamp: timeutil.UnixMilli(timestamp),
		}

		// increment bytes written
		bytesWritten += len(avroData)
	}

	// update log payload
	payload.RecordsCount = len(records)
	payload.BytesWritten = bytesWritten
	middleware.SetTagsPayload(r.Context(), payload)

	// write request to engine
	writeID := engine.GenerateWriteID()
	err = hub.Engine.QueueWriteRequest(r.Context(), &pb_engine.WriteRequest{
		WriteId:    writeID,
		InstanceId: instanceID.Bytes(),
		Records:    records,
	})
	if err != nil {
		return httputil.NewError(400, err.Error())
	}

	// track write metrics
	gateway.Metrics.TrackWrite(instanceID, int64(len(records)), int64(bytesWritten))
	if !secret.IsAnonymous() {
		gateway.Metrics.TrackWrite(secret.GetOwnerID(), int64(len(records)), int64(bytesWritten))
	}

	// Done
	return nil
}
