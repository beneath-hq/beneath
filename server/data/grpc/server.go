package grpc

import (
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	pb "github.com/beneath-hq/beneath/server/data/grpc/proto"
	"github.com/beneath-hq/beneath/services/data"
	"github.com/beneath-hq/beneath/services/middleware"

	// see https://github.com/grpc/grpc-go/blob/master/Documentation/encoding.md#using-a-compressor
	_ "google.golang.org/grpc/encoding/gzip"
)

const (
	maxRecvMsgSize = 1024 * 1024 * 10
	maxSendMsgSize = 1024 * 1024 * 50
)

var kaep = keepalive.EnforcementPolicy{
	MinTime:             30 * time.Second, // If a client pings more than once every 5 seconds, terminate the connection
	PermitWithoutStream: true,             // Allow pings even when there are no active streams
}

var kasp = keepalive.ServerParameters{
	MaxConnectionIdle:     50 * time.Second,
	MaxConnectionAge:      30 * time.Minute,
	MaxConnectionAgeGrace: 1 * time.Minute,
	// Time:                  5 * time.Second,
	// Timeout:               1 * time.Second,
}

// gRPCServer implements pb.GatewayServer
type gRPCServer struct {
	Service *data.Service
}

// NewServer creates and returns the data GRPC server
func NewServer(logger *zap.Logger, data *data.Service, middleware *middleware.Service) *grpc.Server {
	l := logger.Named("grpc")
	server := grpc.NewServer(
		grpc.MaxRecvMsgSize(maxRecvMsgSize),
		grpc.MaxSendMsgSize(maxSendMsgSize),
		grpc.KeepaliveEnforcementPolicy(kaep),
		grpc.KeepaliveParams(kasp),
		grpc_middleware.WithUnaryServerChain(
			middleware.InjectTagsUnaryServerInterceptor(),
			middleware.LoggerUnaryServerInterceptor(l),
			grpc_auth.UnaryServerInterceptor(middleware.AuthInterceptor),
			middleware.RecovererUnaryServerInterceptor(l),
		),
		grpc_middleware.WithStreamServerChain(
			middleware.InjectTagsStreamServerInterceptor(),
			middleware.LoggerStreamServerInterceptor(l),
			grpc_auth.StreamServerInterceptor(middleware.AuthInterceptor),
			middleware.RecovererStreamServerInterceptor(l),
		),
	)
	pb.RegisterGatewayServer(server, &gRPCServer{
		Service: data,
	})
	return server
}
