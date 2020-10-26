package grpc

import (
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"google.golang.org/grpc"

	pb "gitlab.com/beneath-hq/beneath/server/data/grpc/proto"
	"gitlab.com/beneath-hq/beneath/services/data"
	"gitlab.com/beneath-hq/beneath/services/middleware"

	// see https://github.com/grpc/grpc-go/blob/master/Documentation/encoding.md#using-a-compressor
	_ "google.golang.org/grpc/encoding/gzip"
)

const (
	maxRecvMsgSize = 1024 * 1024 * 10
	maxSendMsgSize = 1024 * 1024 * 50
)

// gRPCServer implements pb.GatewayServer
type gRPCServer struct {
	Service *data.Service
}

// NewServer creates and returns the data GRPC server
func NewServer(data *data.Service, middleware *middleware.Service) *grpc.Server {
	server := grpc.NewServer(
		grpc.MaxRecvMsgSize(maxRecvMsgSize),
		grpc.MaxSendMsgSize(maxSendMsgSize),
		grpc_middleware.WithUnaryServerChain(
			middleware.InjectTagsUnaryServerInterceptor(),
			middleware.LoggerUnaryServerInterceptor(),
			grpc_auth.UnaryServerInterceptor(middleware.AuthInterceptor),
			middleware.RecovererUnaryServerInterceptor(),
		),
		grpc_middleware.WithStreamServerChain(
			middleware.InjectTagsStreamServerInterceptor(),
			middleware.LoggerStreamServerInterceptor(),
			grpc_auth.StreamServerInterceptor(middleware.AuthInterceptor),
			middleware.RecovererStreamServerInterceptor(),
		),
	)
	pb.RegisterGatewayServer(server, &gRPCServer{
		Service: data,
	})
	return server
}
