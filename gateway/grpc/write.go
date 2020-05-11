package grpc

import (
	"context"
	"time"

	"gitlab.com/beneath-hq/beneath/gateway/util"

	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"gitlab.com/beneath-hq/beneath/control/entity"
	"gitlab.com/beneath-hq/beneath/engine"
	pb_engine "gitlab.com/beneath-hq/beneath/engine/proto"
	pb "gitlab.com/beneath-hq/beneath/gateway/grpc/proto"
	"gitlab.com/beneath-hq/beneath/internal/hub"
	"gitlab.com/beneath-hq/beneath/internal/middleware"
	"gitlab.com/beneath-hq/beneath/pkg/timeutil"
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
	if stream.Final {
		return nil, status.Error(codes.FailedPrecondition, "instance closed for further writes because it has been marked final")
	}

	// check permissions
	perms := secret.StreamPermissions(ctx, stream.StreamID, stream.ProjectID, stream.Public)
	if !perms.Write {
		return nil, grpc.Errorf(codes.PermissionDenied, "secret doesn't grant right to write to this stream")
	}

	// check records supplied
	if len(req.Records) == 0 {
		return nil, status.Error(codes.InvalidArgument, "records cannot be empty")
	}

	// check quota
	err := util.CheckWriteQuota(ctx, secret)
	if err != nil {
		return nil, status.Error(codes.ResourceExhausted, err.Error())
	}

	// check the batch length is valid
	err = hub.Engine.CheckBatchLength(len(req.Records))
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

		err = hub.Engine.CheckRecordSize(stream, structured, len(record.AvroData))
		if err != nil {
			return nil, grpc.Errorf(codes.InvalidArgument, "error for record at index %d: %v", idx, err.Error())
		}

		// increment bytes written
		bytesWritten += len(record.AvroData)
	}

	// write request to engine
	writeID := engine.GenerateWriteID()
	err = hub.Engine.QueueWriteRequest(ctx, &pb_engine.WriteRequest{
		WriteId:    writeID,
		InstanceId: req.InstanceId,
		Records:    req.Records,
	})
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	// track write metrics
	util.TrackWrite(ctx, secret, stream.StreamID, instanceID, int64(len(req.Records)), int64(bytesWritten))

	// update log payload
	payload.BytesWritten = bytesWritten
	middleware.SetTagsPayload(ctx, payload)

	response := &pb.WriteResponse{
		WriteId: writeID,
	}
	return response, nil
}
