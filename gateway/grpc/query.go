package grpc

import (
	"context"

	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"gitlab.com/beneath-hq/beneath/control/entity"
	pb "gitlab.com/beneath-hq/beneath/gateway/grpc/proto"
	"gitlab.com/beneath-hq/beneath/gateway/util"
	"gitlab.com/beneath-hq/beneath/internal/hub"
	"gitlab.com/beneath-hq/beneath/internal/middleware"
	"gitlab.com/beneath-hq/beneath/pkg/queryparse"
)

func (s *gRPCServer) QueryLog(ctx context.Context, req *pb.QueryLogRequest) (*pb.QueryLogResponse, error) {
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
	payload := queryLogTags{
		InstanceID: instanceID,
		Partitions: req.Partitions,
		Peek:       req.Peek,
	}
	middleware.SetTagsPayload(ctx, payload)

	// get cached stream
	stream := entity.FindCachedStreamByCurrentInstanceID(ctx, instanceID)
	if stream == nil {
		return nil, status.Error(codes.NotFound, "stream not found")
	}

	// check permissions
	perms := secret.StreamPermissions(ctx, stream.StreamID, stream.ProjectID, stream.Public)
	if !perms.Read {
		return nil, grpc.Errorf(codes.PermissionDenied, "token doesn't grant right to read this stream")
	}

	// run peek query
	if req.Peek {
		// check partitions == 1 on peek
		if req.Partitions > 1 {
			return nil, grpc.Errorf(codes.InvalidArgument, "cannot return more than one partition for a peek")
		}

		// run query
		replayCursor, changeCursor, err := hub.Engine.Lookup.Peek(ctx, stream, stream, entity.EfficientStreamInstance(instanceID))
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "error parsing query: %s", err.Error())
		}

		// done
		return &pb.QueryLogResponse{
			ReplayCursors: [][]byte{wrapCursor(util.LogCursorType, instanceID, replayCursor)},
			ChangeCursors: [][]byte{wrapCursor(util.LogCursorType, instanceID, changeCursor)},
		}, nil
	}

	// run normal query
	replayCursors, changeCursors, err := hub.Engine.Lookup.ParseQuery(ctx, stream, stream, entity.EfficientStreamInstance(instanceID), nil, false, int(req.Partitions))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "error parsing query: %s", err.Error())
	}

	// done
	return &pb.QueryLogResponse{
		ReplayCursors: wrapCursors(util.LogCursorType, instanceID, replayCursors),
		ChangeCursors: wrapCursors(util.LogCursorType, instanceID, changeCursors),
	}, nil
}

func (s *gRPCServer) QueryIndex(ctx context.Context, req *pb.QueryIndexRequest) (*pb.QueryIndexResponse, error) {
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
	payload := queryIndexTags{
		InstanceID: instanceID,
		Partitions: req.Partitions,
		Filter:     req.Filter,
	}
	middleware.SetTagsPayload(ctx, payload)

	// get cached stream
	stream := entity.FindCachedStreamByCurrentInstanceID(ctx, instanceID)
	if stream == nil {
		return nil, status.Error(codes.NotFound, "stream not found")
	}

	// check permissions
	perms := secret.StreamPermissions(ctx, stream.StreamID, stream.ProjectID, stream.Public)
	if !perms.Read {
		return nil, grpc.Errorf(codes.PermissionDenied, "token doesn't grant right to read this stream")
	}

	// get filter
	where, err := queryparse.JSONStringToQuery(req.Filter)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "couldn't parse 'where': %s", err.Error())
	}

	// run query
	replayCursors, changeCursors, err := hub.Engine.Lookup.ParseQuery(ctx, stream, stream, entity.EfficientStreamInstance(instanceID), where, true, int(req.Partitions))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "error parsing query: %s", err.Error())
	}

	// done
	return &pb.QueryIndexResponse{
		ReplayCursors: wrapCursors(util.IndexCursorType, instanceID, replayCursors),
		ChangeCursors: wrapCursors(util.LogCursorType, instanceID, changeCursors),
	}, nil
}

func (s *gRPCServer) QueryWarehouse(ctx context.Context, req *pb.QueryWarehouseRequest) (*pb.QueryWarehouseResponse, error) {
	return nil, grpc.Errorf(codes.Unimplemented, "QueryWarehouse is not yet implemented")
}

func (s *gRPCServer) PollWarehouseJob(ctx context.Context, req *pb.PollWarehouseJobRequest) (*pb.PollWarehouseJobResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "PollWarehouseJob is not yet implemented")
}

func wrapCursor(cursorType util.CursorType, id uuid.UUID, engineCursor []byte) []byte {
	if len(engineCursor) == 0 {
		return engineCursor
	}
	return util.NewCursor(cursorType, id, engineCursor).GetBytes()
}

func wrapCursors(cursorType util.CursorType, id uuid.UUID, engineCursors [][]byte) [][]byte {
	wrapped := make([][]byte, len(engineCursors))
	for idx, engineCursor := range engineCursors {
		wrapped[idx] = wrapCursor(cursorType, id, engineCursor)
	}
	return wrapped
}
