package grpc

import (
	"context"

	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"gitlab.com/beneath-hq/beneath/gateway/api"
	pb "gitlab.com/beneath-hq/beneath/gateway/grpc/proto"
)

func (s *gRPCServer) QueryLog(ctx context.Context, req *pb.QueryLogRequest) (*pb.QueryLogResponse, error) {
	// read instanceID
	instanceID := uuid.FromBytesOrNil(req.InstanceId)
	if instanceID == uuid.Nil {
		return nil, status.Error(codes.InvalidArgument, "instance_id not valid UUID")
	}

	// handle
	res, err := api.HandleQueryLog(ctx, &api.QueryLogRequest{
		InstanceID: instanceID,
		Partitions: req.Partitions,
		Peek:       req.Peek,
	})
	if err != nil {
		return nil, err.GRPC()
	}

	// done
	return &pb.QueryLogResponse{
		ReplayCursors: res.ReplayCursors,
		ChangeCursors: res.ChangeCursors,
	}, nil
}

func (s *gRPCServer) QueryIndex(ctx context.Context, req *pb.QueryIndexRequest) (*pb.QueryIndexResponse, error) {
	// read instanceID
	instanceID := uuid.FromBytesOrNil(req.InstanceId)
	if instanceID == uuid.Nil {
		return nil, status.Error(codes.InvalidArgument, "instance_id not valid UUID")
	}

	// handle
	res, err := api.HandleQueryIndex(ctx, &api.QueryIndexRequest{
		InstanceID: instanceID,
		Partitions: req.Partitions,
		Filter:     req.Filter,
	})
	if err != nil {
		return nil, err.GRPC()
	}

	// done
	return &pb.QueryIndexResponse{
		ReplayCursors: res.ReplayCursors,
		ChangeCursors: res.ChangeCursors,
	}, nil
}
