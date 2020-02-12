package grpc

import (
	"context"
	"time"

	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/beneath-core/control/entity"
	"github.com/beneath-core/pkg/middleware"
	"github.com/beneath-core/pkg/timeutil"
	"github.com/beneath-core/db"
	"github.com/beneath-core/engine"
	pb_engine "github.com/beneath-core/engine/proto"
	"github.com/beneath-core/gateway"
	pb "github.com/beneath-core/gateway/grpc/proto"
)

func (s *gRPCServer) Write(ctx context.Context, req *pb.WriteRequest) (*pb.WriteResponse, error) {
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
	payload := writeTags{
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
	writeID := engine.GenerateWriteID()
	err = db.Engine.QueueWriteRequest(ctx, &pb_engine.WriteRequest{
		WriteId:    writeID,
		InstanceId: req.InstanceId,
		Records:    req.Records,
	})
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	// track write metrics
	gateway.Metrics.TrackWrite(instanceID, int64(len(req.Records)), int64(bytesWritten))
	gateway.Metrics.TrackWrite(secret.GetOwnerID(), int64(len(req.Records)), int64(bytesWritten))

	// update log payload
	payload.BytesWritten = bytesWritten
	middleware.SetTagsPayload(ctx, payload)

	response := &pb.WriteResponse{
		WriteId: writeID,
	}
	return response, nil
}
