package gateway

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/beneath-core/beneath-go/control/auth"
	"github.com/beneath-core/beneath-go/control/model"
	pb "github.com/beneath-core/beneath-go/proto"
	uuid "github.com/satori/go.uuid"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"

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
			grpc_auth.UnaryServerInterceptor(auth.GRPCInterceptor),
			grpc_recovery.UnaryServerInterceptor(),
		),
		grpc_middleware.WithStreamServerChain(
			grpc_auth.StreamServerInterceptor(auth.GRPCInterceptor),
			grpc_recovery.StreamServerInterceptor(),
		),
	)
	pb.RegisterGatewayServer(server, &gRPCServer{})

	log.Printf("gRPC server running on port %d\n", port)
	return server.Serve(lis)
}

// gRPCServer implements pb.GatewayServer
type gRPCServer struct{}

func (s *gRPCServer) GetStreamDetails(ctx context.Context, req *pb.StreamDetailsRequest) (*pb.StreamDetailsResponse, error) {
	// get auth
	key := auth.GetKey(ctx)

	// get instance ID
	instanceID := model.FindInstanceIDByNameAndProject(req.StreamName, req.ProjectName)
	if instanceID == uuid.Nil {
		return nil, status.Error(codes.NotFound, "stream not found")
	}

	// get stream details
	stream := model.FindCachedStreamByCurrentInstanceID(instanceID)
	if stream == nil {
		return nil, status.Error(codes.NotFound, "stream not found")
	}

	// check permissions
	if !key.ReadsProject(stream.ProjectID) {
		return nil, grpc.Errorf(codes.PermissionDenied, "token doesn't grant right to read this stream")
	}

	// return
	return &pb.StreamDetailsResponse{
		CurrentInstanceId: instanceID.Bytes(),
		ProjectId:         stream.ProjectID.Bytes(),
		ProjectName:       stream.ProjectName,
		StreamName:        stream.StreamName,
		KeyFields:         stream.KeyCodec.GetKeyFields(),
		AvroSchema:        stream.AvroCodec.GetSchemaString(),
		Public:            stream.Public,
		External:          stream.External,
		Batch:             stream.Batch,
		Manual:            stream.Manual,
	}, nil
}

func (s *gRPCServer) WriteRecords(ctx context.Context, req *pb.WriteRecordsRequest) (*pb.WriteRecordsResponse, error) {
	// get auth
	key := auth.GetKey(ctx)

	// read instanceID
	instanceID := uuid.FromBytesOrNil(req.InstanceId)
	if instanceID == uuid.Nil {
		return nil, status.Error(codes.InvalidArgument, "instance_id not valid UUID")
	}

	// get stream info
	stream := model.FindCachedStreamByCurrentInstanceID(instanceID)
	if stream == nil {
		return nil, status.Error(codes.NotFound, "stream not found")
	}

	// check permissions
	if !key.WritesStream(stream) {
		return nil, grpc.Errorf(codes.PermissionDenied, "token doesn't grant right to write to this stream")
	}

	// check records supplied
	if len(req.Records) == 0 {
		return nil, status.Error(codes.InvalidArgument, "records cannot be empty")
	}

	// check each record is valid
	for idx, record := range req.Records {
		// check sequence number
		if err := Engine.CheckSequenceNumber(record.SequenceNumber); err != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("record at index %d: %v", idx, err.Error()))
		}

		// check it decodes
		decodedData, err := stream.AvroCodec.Unmarshal(record.AvroData, false)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("record at index %d doesn't decode: %v", idx, err.Error()))
		}

		// compute key (only used to check size)
		keyData, err := stream.KeyCodec.Marshal(decodedData.(map[string]interface{}))
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("error encoding record at index %d: %v", idx, err.Error()))
		}

		// check size
		err = Engine.CheckSize(len(keyData), len(record.AvroData))
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("error encoding record at index %d: %v", idx, err.Error()))
		}
	}

	// write request to engine
	err := Engine.Streams.QueueWriteRequest(req)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	return &pb.WriteRecordsResponse{}, nil
}
