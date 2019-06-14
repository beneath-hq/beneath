package gateway

import (
	"context"

	"github.com/beneath-core/beneath-gateway/beneath/proto"
)

// GRPCServer implements proto.GatewayServer
type GRPCServer struct{}

// WriteRecord implements hello.GatewayServer
func (s *GRPCServer) WriteRecord(ctx context.Context, in *proto.Record) (*proto.WriteRecordResponse, error) {
	return &proto.WriteRecordResponse{}, nil
}
