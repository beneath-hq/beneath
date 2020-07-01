package grpc

import (
	"context"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"gitlab.com/beneath-hq/beneath/gateway/util"

	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"gitlab.com/beneath-hq/beneath/control/entity"
	"gitlab.com/beneath-hq/beneath/engine"
	pb_engine "gitlab.com/beneath-hq/beneath/engine/proto"
	pb "gitlab.com/beneath-hq/beneath/gateway/grpc/proto"
	"gitlab.com/beneath-hq/beneath/internal/hub"
	"gitlab.com/beneath-hq/beneath/internal/middleware"
	"gitlab.com/beneath-hq/beneath/pkg/timeutil"
)

func (s *gRPCServer) Write(ctx context.Context, req *pb.WriteRequest) (*pb.WriteResponse, error) {
	// get auth
	secret := middleware.GetSecret(ctx)
	if secret == nil {
		return nil, grpc.Errorf(codes.PermissionDenied, "not authenticated")
	}

	// check quota
	err := util.CheckWriteQuota(ctx, secret)
	if err != nil {
		return nil, status.Error(codes.ResourceExhausted, err.Error())
	}

	// check number of instances
	if len(req.InstanceRecords) > maxInstancesPerWrite {
		return nil, grpc.Errorf(codes.InvalidArgument, "a single request cannot write to more than %d instances", maxInstancesPerWrite)
	}

	// get instances concurrently
	mu := &sync.Mutex{}
	instances := make(map[uuid.UUID]*entity.CachedStream, len(req.InstanceRecords))
	group, cctx := errgroup.WithContext(ctx)
	for idx := range req.InstanceRecords {
		ir := req.InstanceRecords[idx] // see https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		group.Go(func() error {
			// parse instance IID
			instanceID := uuid.FromBytesOrNil(ir.InstanceId)
			if instanceID == uuid.Nil {
				return status.Errorf(codes.InvalidArgument, "instance_id is not a valid UUID")
			}

			// check records supplied
			if len(ir.Records) == 0 {
				return status.Errorf(codes.InvalidArgument, "records are empty for instance_id=%s", instanceID.String())
			}

			// get stream info
			stream := entity.FindCachedStreamByCurrentInstanceID(cctx, instanceID)
			if stream == nil {
				return status.Errorf(codes.NotFound, "stream not found for instance_id=%s", instanceID.String())
			}

			// check not already a committed batch stream
			if stream.Final {
				return status.Errorf(codes.FailedPrecondition, "instance_id=%s closed for further writes because it has been marked final", instanceID.String())
			}

			// check permissions
			perms := secret.StreamPermissions(cctx, stream.StreamID, stream.ProjectID, stream.Public)
			if !perms.Write {
				return grpc.Errorf(codes.PermissionDenied, "secret doesn't grant right to write to stream_id=%s", stream.StreamID.String())
			}

			// set stream
			mu.Lock()
			if instances[instanceID] != nil {
				mu.Unlock()
				return status.Errorf(codes.InvalidArgument, "found two InstanceRecords for instance_id=%s", instanceID.String())
			}
			instances[instanceID] = stream
			mu.Unlock()

			return nil
		})
	}

	// run wait group
	err = group.Wait()
	if err != nil {
		return nil, err
	}

	// process each InstanceRecords
	instanceMetrics := make([]writeMetrics, len(req.InstanceRecords))
	for i, ir := range req.InstanceRecords {
		// check the batch length is valid
		err = hub.Engine.CheckBatchLength(len(ir.Records))
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		// find stream
		instanceID := uuid.FromBytesOrNil(ir.InstanceId)
		stream := instances[instanceID]

		// validate each record
		irBytes := 0
		for j, record := range ir.Records {
			// set timestamp to current timestamp if it's 0
			if record.Timestamp == 0 {
				record.Timestamp = timeutil.UnixMilli(time.Now())
			}

			// check it decodes
			structured, err := stream.Codec.UnmarshalAvro(record.AvroData)
			if err != nil {
				return nil, grpc.Errorf(codes.InvalidArgument, "error for instance_id=%s, record at index %d: %v", instanceID.String(), j, err.Error())
			}

			// check record size
			err = hub.Engine.CheckRecordSize(stream, structured, len(record.AvroData))
			if err != nil {
				return nil, grpc.Errorf(codes.InvalidArgument, "error for instance_id=%s, record at index %d: %v", instanceID.String(), j, err.Error())
			}

			// increment bytes written
			irBytes += len(record.AvroData)
		}

		// metrics
		instanceMetrics[i] = writeMetrics{
			InstanceID:   instanceID,
			BytesWritten: irBytes,
			RecordsCount: len(ir.Records),
		}
	}

	// write request to engine
	writeID := engine.GenerateWriteID()
	err = hub.Engine.QueueWriteRequest(ctx, &pb_engine.WriteRequest{
		WriteId:         writeID,
		InstanceRecords: req.InstanceRecords,
	})
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	// track write metrics
	for _, im := range instanceMetrics {
		stream := instances[im.InstanceID]
		util.TrackWrite(ctx, secret, stream.StreamID, im.InstanceID, int64(im.RecordsCount), int64(im.BytesWritten))
	}

	// set log payload
	payload := writeTags{
		WriteID:   writeID,
		Instances: instanceMetrics,
	}
	middleware.SetTagsPayload(ctx, payload)

	response := &pb.WriteResponse{
		WriteId: writeID,
	}
	return response, nil
}
