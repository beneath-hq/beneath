package grpc

import (
	"context"
	"fmt"

	"github.com/beneath-core/beneath-go/control/entity"
	"github.com/beneath-core/beneath-go/core/jsonutil"
	"github.com/beneath-core/beneath-go/core/middleware"
	"github.com/beneath-core/beneath-go/core/queryparse"
	"github.com/beneath-core/beneath-go/gateway"
	pb "github.com/beneath-core/beneath-go/proto"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	defaultLookupLimit = 50
	maxLookupLimit     = 1000
)

func (s *gRPCServer) LookupRecords(ctx context.Context, req *pb.LookupRecordsRequest) (*pb.LookupRecordsResponse, error) {
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

	// set log payload
	middleware.SetTagsPayload(ctx, lokupRecordsLog{
		InstanceID: instanceID.String(),
		Limit:      req.Limit,
		Where:      req.Where,
		After:      req.After,
	})

	// get cached stream
	stream := entity.FindCachedStreamByCurrentInstanceID(ctx, instanceID)
	if stream == nil {
		return nil, status.Error(codes.NotFound, "stream not found")
	}

	// check permissions
	perms := secret.StreamPermissions(ctx, stream.StreamID, stream.ProjectID, stream.External, stream.Public)
	if !perms.Read {
		return nil, grpc.Errorf(codes.PermissionDenied, "token doesn't grant right to read this stream")
	}

	// check quota
	usage := gateway.Metrics.GetCurrentUsage(ctx, secret.GetOwnerID())
	ok := secret.CheckReadQuota(usage)
	if !ok {
		return nil, status.Error(codes.ResourceExhausted, "you have exhausted your monthly quota")
	}

	// check limit is valid
	if req.Limit == 0 {
		return nil, grpc.Errorf(codes.InvalidArgument, "limit cannot be 0")
	} else if req.Limit > maxLookupLimit {
		return nil, grpc.Errorf(codes.InvalidArgument, fmt.Sprintf("limit exceeds maximum of %d", maxLookupLimit))
	}

	// get key range where clause (if no where, it will be nil)
	var whereQuery queryparse.Query
	if req.Where != "" {
		// read where
		var where map[string]interface{}
		err := jsonutil.UnmarshalBytes([]byte(req.Where), &where)
		if err != nil {
			return nil, grpc.Errorf(codes.InvalidArgument, "couldn't parse 'where' -- is it valid JSON?")
		}

		// make query
		whereQuery, err = queryparse.JSONToQuery(where)
		if err != nil {
			return nil, grpc.Errorf(codes.InvalidArgument, fmt.Sprintf("couldn't parse 'where' query: %s", err.Error()))
		}
	}

	// adapt key range based on after (for pagination), if present
	var afterQuery queryparse.Query
	if req.After != "" {
		// read after
		var after map[string]interface{}
		err := jsonutil.UnmarshalBytes([]byte(req.After), &after)
		if err != nil {
			return nil, grpc.Errorf(codes.InvalidArgument, "couldn't parse 'after' -- is it valid JSON?")
		}

		// make query
		afterQuery, err = queryparse.JSONToQuery(after)
		if err != nil {
			return nil, grpc.Errorf(codes.InvalidArgument, fmt.Sprintf("couldn't parse 'after' query: %s", err.Error()))
		}
	}

	// set key range
	_, err := stream.Codec.MakeKeyRange(whereQuery, afterQuery)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	// read rows from engine
	response := &pb.LookupRecordsResponse{}
	bytesRead := 0

	// TODO
	// err = db.Engine.Tables.ReadRecordRange(ctx, instanceID, keyRange, int(req.Limit), func(avroData []byte, timestamp time.Time) error {
	// 	response.Records = append(response.Records, &pb.Record{
	// 		AvroData:  avroData,
	// 		Timestamp: timeutil.UnixMilli(timestamp),
	// 	})
	// 	bytesRead += len(avroData)
	// 	return nil
	// })
	// if err != nil {
	// 	return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	// }

	// track read metrics
	gateway.Metrics.TrackRead(stream.StreamID, int64(len(response.Records)), int64(bytesRead))
	gateway.Metrics.TrackRead(secret.GetOwnerID(), int64(len(response.Records)), int64(bytesRead))

	// done
	return response, nil
}
