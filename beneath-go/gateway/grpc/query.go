package grpc

import (
	"context"

	"github.com/beneath-core/beneath-go/control/entity"
	"github.com/beneath-core/beneath-go/core/middleware"
	"github.com/beneath-core/beneath-go/core/queryparse"
	"github.com/beneath-core/beneath-go/core/timeutil"
	"github.com/beneath-core/beneath-go/db"
	"github.com/beneath-core/beneath-go/gateway"
	pb "github.com/beneath-core/beneath-go/proto"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *gRPCServer) Peek(ctx context.Context, req *pb.PeekRequest) (*pb.PeekResponse, error) {
	panic("not implemented")
}

func (s *gRPCServer) Subscribe(req *pb.SubscribeRequest, ss pb.Gateway_SubscribeServer) error {
	panic("not implemented")
}

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
	replayCursors, changeCursors, err := db.Engine.Query(ctx, stream, stream, stream, where, req.Compact, int(req.Partitions))
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
	it, err := db.Engine.ReadCursor(ctx, stream, stream, stream, req.Cursor, int(req.Limit))
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "%s", err.Error())
	}

	// make response
	response := &pb.ReadResponse{}
	bytesRead := 0

	for {
		record := it.Next()
		if record == nil {
			break
		}

		recordProto := &pb.Record{
			AvroData:  record.GetAvro(),
			Timestamp: timeutil.UnixMilli(record.GetTimestamp()),
		}

		bytesRead += len(recordProto.AvroData)
		response.Records = append(response.Records, recordProto)
	}

	// set next cursor
	response.NextCursor = it.NextPageCursor()

	// track read metrics
	gateway.Metrics.TrackRead(stream.StreamID, int64(len(response.Records)), int64(bytesRead))
	gateway.Metrics.TrackRead(secret.GetOwnerID(), int64(len(response.Records)), int64(bytesRead))

	// update log message
	payload.BytesRead = bytesRead
	middleware.SetTagsPayload(ctx, payload)

	// done
	return response, nil
}
