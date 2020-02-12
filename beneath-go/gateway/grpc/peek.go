package grpc

import (
	"context"

	pb "github.com/beneath-core/beneath-go/gateway/grpc/proto"
)

func (s *gRPCServer) Peek(ctx context.Context, req *pb.PeekRequest) (*pb.PeekResponse, error) {
	panic("not implemented")
}
