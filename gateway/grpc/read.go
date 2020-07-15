package grpc

import (
	"context"
	"fmt"

	"gitlab.com/beneath-hq/beneath/engine/driver"

	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"gitlab.com/beneath-hq/beneath/control/entity"
	pb "gitlab.com/beneath-hq/beneath/gateway/grpc/proto"
	"gitlab.com/beneath-hq/beneath/gateway/util"
	"gitlab.com/beneath-hq/beneath/internal/hub"
	"gitlab.com/beneath-hq/beneath/internal/middleware"
	"gitlab.com/beneath-hq/beneath/pkg/timeutil"
)

func (s *gRPCServer) Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error) {
	// get auth
	secret := middleware.GetSecret(ctx)
	if secret == nil {
		return nil, grpc.Errorf(codes.PermissionDenied, "not authenticated")
	}

	// set payload
	payload := readTags{
		Cursor: req.Cursor,
		Limit:  req.Limit,
	}
	middleware.SetTagsPayload(ctx, payload)

	// parse cursor
	cursor, err := util.CursorFromBytes(req.Cursor)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "%s", err.Error())
	}

	// check limit
	if req.Limit == 0 {
		req.Limit = defaultReadLimit
	} else if req.Limit > maxReadLimit {
		return nil, grpc.Errorf(codes.InvalidArgument, "limit exceeds maximum (%d)", maxReadLimit)
	}

	// check quota
	err = util.CheckReadQuota(ctx, secret)
	if err != nil {
		return nil, status.Error(codes.ResourceExhausted, err.Error())
	}

	// get result iterator
	var it driver.RecordsIterator
	streamID := uuid.Nil
	instanceID := uuid.Nil

	if cursor.GetType() == util.LogCursorType || cursor.GetType() == util.IndexCursorType {
		// get instanceID
		instanceID = cursor.GetID()

		// update tags
		payload.InstanceID = &instanceID
		middleware.SetTagsPayload(ctx, payload)

		// get cached stream
		stream := entity.FindCachedStreamByCurrentInstanceID(ctx, instanceID)
		if stream == nil {
			return nil, status.Error(codes.NotFound, "stream not found")
		}
		streamID = stream.StreamID

		// check permissions
		perms := secret.StreamPermissions(ctx, stream.StreamID, stream.ProjectID, stream.Public)
		if !perms.Read {
			return nil, grpc.Errorf(codes.PermissionDenied, "token doesn't grant right to read this stream")
		}

		// get it
		it, err = hub.Engine.Lookup.ReadCursor(ctx, stream, stream, entity.EfficientStreamInstance(instanceID), cursor.GetPayload(), int(req.Limit))
		if err != nil {
			return nil, grpc.Errorf(codes.InvalidArgument, "%s", err.Error())
		}
	} else if cursor.GetType() == util.WarehouseCursorType {
		// get jobID
		jobID := cursor.GetID()

		// update tags
		payload.JobID = &jobID
		middleware.SetTagsPayload(ctx, payload)

		// get it
		it, err = hub.Engine.Warehouse.ReadWarehouseCursor(ctx, cursor.GetPayload(), int(req.Limit))
		if err != nil {
			return nil, grpc.Errorf(codes.InvalidArgument, "%s", err.Error())
		}
	} else {
		panic(fmt.Errorf("cannot handle cursor type"))
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
	response.NextCursor = wrapCursor(cursor.GetType(), cursor.GetID(), it.NextCursor())

	// track read metrics
	util.TrackRead(ctx, secret, streamID, instanceID, int64(len(response.Records)), int64(bytesRead))

	// update log message
	payload.BytesRead = bytesRead
	middleware.SetTagsPayload(ctx, payload)

	// done
	return response, nil
}
