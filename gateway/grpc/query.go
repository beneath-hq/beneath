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

func (s *gRPCServer) Repartition(ctx context.Context, req *pb.RepartitionRequest) (*pb.RepartitionResponse, error) {
	panic("not implemented")
}

func (s *gRPCServer) Query(ctx context.Context, req *pb.QueryRequest) (*pb.QueryResponse, error) {
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
	payload := queryTags{
		InstanceID: instanceID.String(),
		Filter:     req.Filter,
		Partitions: req.Partitions,
		Compact:    req.Compact,
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
	replayCursors, changeCursors, err := hub.Engine.Lookup.ParseQuery(ctx, stream, stream, stream, where, req.Compact, int(req.Partitions))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "couldn't parse 'where': %s", err.Error())
	}

	// done
	return &pb.QueryResponse{
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
