package gateway

import (
	"context"
	"fmt"
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"

	pb "github.com/beneath-core/beneath-gateway/beneath/beneath_proto"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ListenAndServeGRPC serves a gRPC API
func ListenAndServeGRPC(port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	server := grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(
			grpc_auth.UnaryServerInterceptor(authInterceptor),
			grpc_recovery.UnaryServerInterceptor(),
		),
		grpc_middleware.WithStreamServerChain(
			grpc_auth.StreamServerInterceptor(authInterceptor),
			grpc_recovery.StreamServerInterceptor(),
		),
	)
	pb.RegisterGatewayServer(server, &gRPCServer{})

	return server.Serve(lis)
}

// gRPCServer implements pb.GatewayServer
type gRPCServer struct{}

func (s *gRPCServer) WriteEncodedRecords(ctx context.Context, req *pb.WriteEncodedRecordsRequest) (*pb.WriteEncodedRecordsResponse, error) {
	auth := getAuth(ctx)

	instanceID, err := uuid.FromBytes(req.InstanceId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "InstanceId not valid UUID")
	}

	instance, err := lookupInstance(instanceID)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	role, err := lookupRole(auth, instance)
	if err != nil {
		return nil, grpc.Errorf(codes.NotFound, err.Error())
	}

	if !role.Write {
		return nil, grpc.Errorf(codes.PermissionDenied, "token doesn't grant right to write to this stream")
	}

	err = engine.QueueWrite(req)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	return &pb.WriteEncodedRecordsResponse{}, nil
}
