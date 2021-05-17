package grpc

import (
	"context"

	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/beneath-hq/beneath/server/data/grpc/proto"
	"github.com/beneath-hq/beneath/services/data"
)

func (s *gRPCServer) Write(ctx context.Context, req *pb.WriteRequest) (*pb.WriteResponse, error) {
	// build instanceRecords
	instanceRecords := make(map[uuid.UUID]*data.WriteRecords)
	for _, ir := range req.InstanceRecords {
		instanceID := uuid.FromBytesOrNil(ir.InstanceId)
		if instanceID == uuid.Nil {
			return nil, status.Errorf(codes.InvalidArgument, "instance_id <%v> is not a valid UUID", ir.InstanceId)
		}
		if instanceRecords[instanceID] != nil {
			return nil, status.Errorf(codes.InvalidArgument, "found two InstanceRecords for instance_id=%s", instanceID.String())
		}
		instanceRecords[instanceID] = &data.WriteRecords{PB: ir.Records}
	}

	// call write
	res, err := s.Service.HandleWrite(ctx, &data.WriteRequest{InstanceRecords: instanceRecords})
	if err != nil {
		return nil, err.GRPC()
	}

	return &pb.WriteResponse{WriteId: res.WriteID.Bytes()}, nil
}
