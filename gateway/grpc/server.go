package grpc

import (
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"google.golang.org/grpc"

	pb "gitlab.com/beneath-org/beneath/gateway/grpc/proto"
	"gitlab.com/beneath-org/beneath/internal/middleware"

	// see https://github.com/grpc/grpc-go/blob/master/Documentation/encoding.md#using-a-compressor
	_ "google.golang.org/grpc/encoding/gzip"
)

const (
	defaultReadLimit = 50
	maxReadLimit     = 1000
)

const (
	maxRecvMsgSize = 1024 * 1024 * 10
	maxSendMsgSize = 1024 * 1024 * 50
)

// gRPCServer implements pb.GatewayServer
type gRPCServer struct{}

// Server returns the gateway GRPC server
func Server() *grpc.Server {
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
	pb.RegisterGatewayServer(server, &gRPCServer{})
	return server
}

type pingTags struct {
	ClientID      string `json:"client_id,omitempty"`
	ClientVersion string `json:"client_version,omitempty"`
}

type writeTags struct {
	InstanceID   string `json:"instance_id,omitempty"`
	RecordsCount int    `json:"records,omitempty"`
	BytesWritten int    `json:"bytes,omitempty"`
}

type queryLogTags struct {
	InstanceID string `json:"instance_id,omitempty"`
	Partitions int32  `json:"partitions,omitempty"`
	Peek       bool   `json:"peek,omitempty"`
}

type queryIndexTags struct {
	InstanceID string `json:"instance_id,omitempty"`
	Partitions int32  `json:"partitions,omitempty"`
	Filter     string `json:"filter,omitempty"`
}

type readTags struct {
	InstanceID string `json:"instance_id,omitempty"`
	Cursor     []byte `json:"cursor,omitempty"`
	Limit      int32  `json:"limit,omitempty"`
	BytesRead  int    `json:"bytes,omitempty"`
}
