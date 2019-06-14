package gateway

import (
	"context"
	"fmt"
	"net"

	"github.com/beneath-core/beneath-gateway/beneath/proto"
	"google.golang.org/grpc"
)

// ListenAndServeGRPC serves a gRPC API
func ListenAndServeGRPC(port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	server := grpc.NewServer()
	proto.RegisterGatewayServer(server, &gRPCServer{})

	// TODO: Authentication

	return server.Serve(lis)
}

// gRPCServer implements proto.GatewayServer
type gRPCServer struct{}

func (s *gRPCServer) WriteRecords(ctx context.Context, in *proto.WriteRecordsRequest) (*proto.WriteRecordResponse, error) {
	return &proto.WriteRecordResponse{}, nil
}
