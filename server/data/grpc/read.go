package grpc

import (
	"context"

	pb "github.com/beneath-hq/beneath/server/data/grpc/proto"
	"github.com/beneath-hq/beneath/services/data"
)

func (s *gRPCServer) Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error) {
	res, err := s.Service.HandleRead(ctx, &data.ReadRequest{
		Cursor:   req.Cursor,
		Limit:    req.Limit,
		ReturnPB: true,
	})
	if err != nil {
		return nil, err.GRPC()
	}

	return &pb.ReadResponse{
		Records:    res.PB,
		NextCursor: res.NextCursor,
	}, nil
}
