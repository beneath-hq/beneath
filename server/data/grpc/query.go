package grpc

import (
	"context"

	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/beneath-hq/beneath/server/data/grpc/proto"
	"github.com/beneath-hq/beneath/services/data"
)

func (s *gRPCServer) QueryLog(ctx context.Context, req *pb.QueryLogRequest) (*pb.QueryLogResponse, error) {
	// read instanceID
	instanceID := uuid.FromBytesOrNil(req.InstanceId)
	if instanceID == uuid.Nil {
		return nil, status.Error(codes.InvalidArgument, "instance_id not valid UUID")
	}

	// handle
	res, err := s.Service.HandleQueryLog(ctx, &data.QueryLogRequest{
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
	res, err := s.Service.HandleQueryIndex(ctx, &data.QueryIndexRequest{
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
