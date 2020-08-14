package grpc

import (
	"context"

	"gitlab.com/beneath-hq/beneath/gateway/api"
	pb "gitlab.com/beneath-hq/beneath/gateway/grpc/proto"
)

func (s *gRPCServer) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	res, err := api.HandlePing(ctx, &api.PingRequest{
		ClientID:      req.ClientId,
		ClientVersion: req.ClientVersion,
	})
	if err != nil {
		return nil, err.GRPC()
	}
	return &pb.PingResponse{
		Authenticated:      res.Authenticated,
		VersionStatus:      res.VersionStatus,
		RecommendedVersion: res.RecommendedVersion,
	}, nil
}
