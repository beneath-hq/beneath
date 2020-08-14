package grpc

import (
	"context"

	"gitlab.com/beneath-hq/beneath/gateway/api"
	pb "gitlab.com/beneath-hq/beneath/gateway/grpc/proto"
)

func (s *gRPCServer) Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error) {
	res, err := api.HandleRead(ctx, &api.ReadRequest{
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
