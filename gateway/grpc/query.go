package grpc

import (
	"context"

	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/beneath-core/control/entity"
	"github.com/beneath-core/gateway"
	pb "github.com/beneath-core/gateway/grpc/proto"
	"github.com/beneath-core/internal/hub"
	"github.com/beneath-core/internal/middleware"
	"github.com/beneath-core/pkg/queryparse"
	"github.com/beneath-core/pkg/timeutil"
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
		InstanceID: instanceID.String(),
		Partitions: req.Partitions,
		Peek:       req.Peek,
	}
	middleware.SetTagsPayload(ctx, payload)

	// get cached stream
	stream := entity.FindCachedStreamByCurrentInstanceID(ctx, instanceID)
	if stream == nil {
		return nil, status.Error(codes.NotFound, "stream not found")
	}

	// if batch, check committed
	if stream.Batch && !stream.Committed {
		return nil, status.Error(codes.FailedPrecondition, "batch has not yet been committed, and so can't be read")
	}

	// check permissions
	perms := secret.StreamPermissions(ctx, stream.StreamID, stream.ProjectID, stream.Public, stream.External)
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
		replayCursor, changeCursor, err := hub.Engine.Lookup.Peek(ctx, stream, stream, stream)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "error parsing query: %s", err.Error())
		}

		// done
		return &pb.QueryLogResponse{
			ReplayCursors: [][]byte{replayCursor},
			ChangeCursors: [][]byte{changeCursor},
		}, nil
	}

	// run normal query
	replayCursors, changeCursors, err := hub.Engine.Lookup.ParseQuery(ctx, stream, stream, stream, nil, false, int(req.Partitions))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "error parsing query: %s", err.Error())
	}

	// done
	return &pb.QueryLogResponse{
		ReplayCursors: replayCursors,
		ChangeCursors: changeCursors,
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
		InstanceID: instanceID.String(),
		Partitions: req.Partitions,
		Filter:     req.Filter,
	}
	middleware.SetTagsPayload(ctx, payload)

	// get cached stream
	stream := entity.FindCachedStreamByCurrentInstanceID(ctx, instanceID)
	if stream == nil {
		return nil, status.Error(codes.NotFound, "stream not found")
	}

	// if batch, check committed
	if stream.Batch && !stream.Committed {
		return nil, status.Error(codes.FailedPrecondition, "batch has not yet been committed, and so can't be read")
	}

	// check permissions
	perms := secret.StreamPermissions(ctx, stream.StreamID, stream.ProjectID, stream.Public, stream.External)
	if !perms.Read {
		return nil, grpc.Errorf(codes.PermissionDenied, "token doesn't grant right to read this stream")
	}

	// get filter
	where, err := queryparse.JSONStringToQuery(req.Filter)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "couldn't parse 'where': %s", err.Error())
	}

	// run query
	replayCursors, changeCursors, err := hub.Engine.Lookup.ParseQuery(ctx, stream, stream, stream, where, true, int(req.Partitions))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "error parsing query: %s", err.Error())
	}

	// done
	return &pb.QueryIndexResponse{
		ReplayCursors: replayCursors,
		ChangeCursors: changeCursors,
	}, nil
}

func (s *gRPCServer) Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error) {
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
	payload := readTags{
		InstanceID: instanceID.String(),
		Cursor:     req.Cursor,
		Limit:      req.Limit,
	}
	middleware.SetTagsPayload(ctx, payload)

	// get cached stream
	stream := entity.FindCachedStreamByCurrentInstanceID(ctx, instanceID)
	if stream == nil {
		return nil, status.Error(codes.NotFound, "stream not found")
	}

	// if batch, check committed
	if stream.Batch && !stream.Committed {
		return nil, status.Error(codes.FailedPrecondition, "batch has not yet been committed, and so can't be read")
	}

	// check permissions
	perms := secret.StreamPermissions(ctx, stream.StreamID, stream.ProjectID, stream.Public, stream.External)
	if !perms.Read {
		return nil, grpc.Errorf(codes.PermissionDenied, "token doesn't grant right to read this stream")
	}

	// check limit
	if req.Limit == 0 {
		req.Limit = defaultReadLimit
	} else if req.Limit > maxReadLimit {
		return nil, grpc.Errorf(codes.InvalidArgument, "limit exceeds maximum (%d)", maxReadLimit)
	}

	// check quota
	usage := gateway.Metrics.GetCurrentUsage(ctx, secret.GetOwnerID())
	ok := secret.CheckReadQuota(usage)
	if !ok {
		return nil, status.Error(codes.ResourceExhausted, "you have exhausted your monthly quota")
	}

	// get result iterator
	it, err := hub.Engine.Lookup.ReadCursor(ctx, stream, stream, stream, req.Cursor, int(req.Limit))
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "%s", err.Error())
	}

	// make response
	response := &pb.ReadResponse{}
	bytesRead := 0

	for it.Next() {
		record := it.Record()

		recordProto := &pb.Record{
			AvroData:  record.GetAvro(),
			Timestamp: timeutil.UnixMilli(record.GetTimestamp()),
		}

		bytesRead += len(recordProto.AvroData)
		response.Records = append(response.Records, recordProto)
	}

	// set next cursor
	response.NextCursor = it.NextCursor()

	// track read metrics
	gateway.Metrics.TrackRead(stream.StreamID, int64(len(response.Records)), int64(bytesRead))
	gateway.Metrics.TrackRead(instanceID, int64(len(response.Records)), int64(bytesRead))
	gateway.Metrics.TrackRead(secret.GetOwnerID(), int64(len(response.Records)), int64(bytesRead))

	// update log message
	payload.BytesRead = bytesRead
	middleware.SetTagsPayload(ctx, payload)

	// done
	return response, nil
}
