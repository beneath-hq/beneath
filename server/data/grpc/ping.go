package grpc

import (
	"context"

	pb "gitlab.com/beneath-hq/beneath/server/data/grpc/proto"
	"gitlab.com/beneath-hq/beneath/services/data"
)

func (s *gRPCServer) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	res, err := s.Service.HandlePing(ctx, &data.PingRequest{
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
