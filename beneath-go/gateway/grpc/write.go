package grpc

import (
	"context"
	"time"

	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/beneath-core/beneath-go/control/entity"
	"github.com/beneath-core/beneath-go/core/middleware"
	"github.com/beneath-core/beneath-go/core/timeutil"
	"github.com/beneath-core/beneath-go/db"
	"github.com/beneath-core/beneath-go/gateway"
	pb "github.com/beneath-core/beneath-go/proto"
)

func (s *gRPCServer) WriteRecords(ctx context.Context, req *pb.WriteRecordsRequest) (*pb.WriteRecordsResponse, error) {
	// get auth
	secret := middleware.GetSecret(ctx)
	if secret == nil {
		return nil, grpc.Errorf(codes.PermissionDenied, "not authenticated")
	}

	// read instanceID
	instanceID := uuid.FromBytesOrNil(req.InstanceId)
	if instanceID == uuid.Nil {
		return nil, status.Error(codes.InvalidArgument, "instance_id not valid UUID")
	}

	// set log payload
	payload := writeRecordsTags{
		InstanceID:   instanceID.String(),
		RecordsCount: len(req.Records),
	}
	middleware.SetTagsPayload(ctx, payload)

	// get stream info
	stream := entity.FindCachedStreamByCurrentInstanceID(ctx, instanceID)
	if stream == nil {
		return nil, status.Error(codes.NotFound, "stream not found")
	}

	// check not already a committed batch stream
	if stream.Batch && stream.Committed {
		return nil, status.Error(codes.FailedPrecondition, "batch has been committed and closed for further writes")
	}

	// check permissions
	perms := secret.StreamPermissions(ctx, stream.StreamID, stream.ProjectID, stream.Public, stream.External)
	if !perms.Write {
		return nil, grpc.Errorf(codes.PermissionDenied, "secret doesn't grant right to write to this stream")
	}

	// check records supplied
	if len(req.Records) == 0 {
		return nil, status.Error(codes.InvalidArgument, "records cannot be empty")
	}

	// check quota
	usage := gateway.Metrics.GetCurrentUsage(ctx, secret.GetOwnerID())
	ok := secret.CheckWriteQuota(usage)
	if !ok {
		return nil, status.Error(codes.ResourceExhausted, "you have exhausted your monthly quota")
	}

	// check the batch length is valid
	err := db.Engine.CheckBatchLength(len(req.Records))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// check each record is valid
	bytesWritten := 0
	for idx, record := range req.Records {
		// set timestamp to current timestamp if it's 0
		if record.Timestamp == 0 {
			record.Timestamp = timeutil.UnixMilli(time.Now())
		}

		// check it decodes
		structured, err := stream.Codec.UnmarshalAvro(record.AvroData)
		if err != nil {
			return nil, grpc.Errorf(codes.InvalidArgument, "error for record at index %d: %v", idx, err.Error())
		}

		err = db.Engine.CheckRecordSize(stream, structured, len(record.AvroData))
		if err != nil {
			return nil, grpc.Errorf(codes.InvalidArgument, "error for record at index %d: %v", idx, err.Error())
		}

		// increment bytes written
		bytesWritten += len(record.AvroData)
	}

	// write request to engine
	err = db.Engine.QueueWriteRequest(ctx, req)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	// track write metrics
	gateway.Metrics.TrackWrite(instanceID, int64(len(req.Records)), int64(bytesWritten))
	gateway.Metrics.TrackWrite(secret.GetOwnerID(), int64(len(req.Records)), int64(bytesWritten))

	// update log payload
	payload.BytesWritten = bytesWritten
	middleware.SetTagsPayload(ctx, payload)

	response := &pb.WriteRecordsResponse{}
	return response, nil
}