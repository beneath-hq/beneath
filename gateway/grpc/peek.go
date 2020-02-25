package grpc

import (
	"context"

	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/beneath-core/control/entity"
	pb "github.com/beneath-core/gateway/grpc/proto"
	"github.com/beneath-core/internal/hub"
	"github.com/beneath-core/internal/middleware"
)

func (s *gRPCServer) Peek(ctx context.Context, req *pb.PeekRequest) (*pb.PeekResponse, error) {
	// get auth
	secret := middleware.GetSecret(ctx)
	if secret == nil {
		return nil, grpc.Errorf(codes.PermissionDenied, "not authenticated")
	}

	// read instanceID
	instanceID := uuid.FromBytesOrNil(req.InstanceId)
	if instanceID == uuid.Nil {
		return nil, status.Error(codes.InvalidArgument, "instance_id not valid UUID")
	}

	// set payload
	payload := peekTags{
		InstanceID: instanceID.String(),
	}
	middleware.SetTagsPayload(ctx, payload)

	// get cached stream
	stream := entity.FindCachedStreamByCurrentInstanceID(ctx, instanceID)
	if stream == nil {
		return nil, status.Error(codes.NotFound, "stream not found")
	}

	// if batch, check committed (sort of silly to peek batch data, but we'll allow it)
	if stream.Batch && !stream.Committed {
		return nil, status.Error(codes.FailedPrecondition, "batch has not yet been committed, and so can't be read")
	}

	// check permissions
	perms := secret.StreamPermissions(ctx, stream.StreamID, stream.ProjectID, stream.Public, stream.External)
	if !perms.Read {
		return nil, grpc.Errorf(codes.PermissionDenied, "token doesn't grant right to read this stream")
	}

	// run query
	rewindCursor, changeCursor, err := hub.Engine.Lookup.Peek(ctx, stream, stream, stream)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "error parsing query: %s", err.Error())
	}

	// done
	return &pb.PeekResponse{
		RewindCursor: rewindCursor,
		ChangeCursor: changeCursor,
	}, nil
}
