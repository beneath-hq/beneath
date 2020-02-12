package grpc

import (
	"context"
	"strings"

	"github.com/beneath-core/control/entity"
	"github.com/beneath-core/internal/middleware"
	pb "github.com/beneath-core/gateway/grpc/proto"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *gRPCServer) GetStream(ctx context.Context, req *pb.GetStreamRequest) (*pb.GetStreamResponse, error) {
	// to backend names
	req.StreamName = toBackendName(req.StreamName)
	req.ProjectName = toBackendName(req.ProjectName)

	// set log payload
	middleware.SetTagsPayload(ctx, streamDetailsTags{
		Stream:  req.StreamName,
		Project: req.ProjectName,
	})

	// get auth
	secret := middleware.GetSecret(ctx)
	if secret == nil {
		return nil, grpc.Errorf(codes.PermissionDenied, "not authenticated")
	}

	// get instance ID
	instanceID := entity.FindInstanceIDByNameAndProject(ctx, req.StreamName, req.ProjectName)
	if instanceID == uuid.Nil {
		return nil, status.Error(codes.NotFound, "stream not found")
	}

	// get stream details
	stream := entity.FindCachedStreamByCurrentInstanceID(ctx, instanceID)
	if stream == nil {
		return nil, status.Error(codes.NotFound, "stream not found")
	}

	// check permissions
	perms := secret.StreamPermissions(ctx, stream.StreamID, stream.ProjectID, stream.Public, stream.External)
	if !perms.Read {
		return nil, grpc.Errorf(codes.PermissionDenied, "token doesn't grant right to read this stream")
	}

	// indexes
	var indexes []*pb.StreamIndexDetails
	// TODO

	// return
	return &pb.GetStreamResponse{
		ProjectId:         stream.ProjectID.Bytes(),
		ProjectName:       stream.ProjectName,
		StreamId:          stream.StreamID.Bytes(),
		StreamName:        stream.StreamName,
		CurrentInstanceId: instanceID.Bytes(),
		Public:            stream.Public,
		External:          stream.External,
		Batch:             stream.Batch,
		Manual:            stream.Manual,
		Committed:         stream.Committed,
		RetentionSeconds:  stream.RetentionSeconds,
		AvroSchema:        stream.Codec.AvroSchemaString,
		Indexes:           indexes,
	}, nil
}

func toBackendName(s string) string {
	return strings.ToLower(strings.ReplaceAll(s, "-", "_"))
}
