package grpc

import (
	"context"

	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"gitlab.com/beneath-hq/beneath/gateway/api"
	pb "gitlab.com/beneath-hq/beneath/gateway/grpc/proto"
)

func (s *gRPCServer) Write(ctx context.Context, req *pb.WriteRequest) (*pb.WriteResponse, error) {
	// build instanceRecords
	instanceRecords := make(map[uuid.UUID]*api.WriteRecords)
	for _, ir := range req.InstanceRecords {
		instanceID := uuid.FromBytesOrNil(ir.InstanceId)
		if instanceID == uuid.Nil {
			return nil, status.Errorf(codes.InvalidArgument, "instance_id <%v> is not a valid UUID", ir.InstanceId)
		}
		if instanceRecords[instanceID] != nil {
			return nil, status.Errorf(codes.InvalidArgument, "found two InstanceRecords for instance_id=%s", instanceID.String())
		}
		instanceRecords[instanceID] = &api.WriteRecords{PB: ir.Records}
	}

	// call write
	res, err := api.HandleWrite(ctx, &api.WriteRequest{InstanceRecords: instanceRecords})
	if err != nil {
		return nil, err.GRPC()
	}

	return &pb.WriteResponse{WriteId: res.WriteID.Bytes()}, nil
}
