package grpc

import (
	"context"
	"fmt"

	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"gitlab.com/beneath-hq/beneath/control/entity"
	"gitlab.com/beneath-hq/beneath/engine/driver"
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
	// get auth
	secret := middleware.GetSecret(ctx)
	if secret == nil {
		return nil, grpc.Errorf(codes.PermissionDenied, "not authenticated")
	}

	// set payload
	payload := queryWarehouseTags{
		Partitions: req.Partitions,
		Query:      req.Query,
		DryRun:     req.DryRun,
		Timeout:    req.TimeoutMs,
		MaxScan:    req.MaxBytesScanned,
	}
	middleware.SetTagsPayload(ctx, payload)

	// expand query
	query, err := hub.Engine.ExpandWarehouseQuery(ctx, req.Query, func(ctx context.Context, organizationName, projectName, streamName string) (driver.Project, driver.Stream, driver.StreamInstance, error) {
		// get instance ID
		instanceID := entity.FindInstanceIDByOrganizationProjectAndName(ctx, organizationName, projectName, streamName)
		if instanceID == uuid.Nil {
			return nil, nil, nil, grpc.Errorf(codes.NotFound, "instance not found for stream '%s/%s/%s'", organizationName, projectName, streamName)
		}

		// get stream
		stream := entity.FindCachedStreamByCurrentInstanceID(ctx, instanceID)
		if stream == nil {
			return nil, nil, nil, grpc.Errorf(codes.NotFound, "stream '%s/%s/%s' not found", organizationName, projectName, streamName)
		}

		// check permissions
		perms := secret.StreamPermissions(ctx, stream.StreamID, stream.ProjectID, stream.Public)
		if !perms.Read {
			return nil, nil, nil, grpc.Errorf(codes.PermissionDenied, "token doesn't grant right to read stream '%s/%s/%s'", organizationName, projectName, streamName)
		}

		// success
		return stream, stream, entity.EfficientStreamInstance(instanceID), nil
	})
	if err != nil {
		if _, ok := status.FromError(err); ok {
			return nil, err
		}
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	// analyze query
	job, err := hub.Engine.Warehouse.AnalyzeWarehouseQuery(ctx, query)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "error during query analysis: %s", err.Error())
	}

	// stop if dry
	if req.DryRun {
		return &pb.QueryWarehouseResponse{
			Job: makeWarehouseJobPB(job),
		}, nil
	}

	// check bytes scanned
	estimatedBytesScanned := job.GetBytesScanned()
	if estimatedBytesScanned > req.MaxBytesScanned {
		return nil, grpc.Errorf(codes.FailedPrecondition, "query would scan %d bytes, which exceeds job limit of %d (limit the query or increase the limit)", job.GetBytesScanned(), req.MaxBytesScanned)
	}

	// check quota
	err = util.CheckScanQuota(ctx, secret, estimatedBytesScanned)
	if err != nil {
		return nil, status.Error(codes.ResourceExhausted, err.Error())
	}

	// run real job
	jobID := uuid.NewV4()
	job, err = hub.Engine.Warehouse.RunWarehouseQuery(ctx, jobID, query, int(req.Partitions), int(req.TimeoutMs), int(req.MaxBytesScanned))
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "error running query: %s", err.Error())
	}

	// update payload
	payload.JobID = &jobID
	middleware.SetTagsPayload(ctx, payload)

	// track read metrics
	util.TrackScan(ctx, secret, estimatedBytesScanned)

	return &pb.QueryWarehouseResponse{
		Job: makeWarehouseJobPB(job),
	}, nil
}

func (s *gRPCServer) PollWarehouseJob(ctx context.Context, req *pb.PollWarehouseJobRequest) (*pb.PollWarehouseJobResponse, error) {
	// read jobID
	jobID := uuid.FromBytesOrNil(req.JobId)
	if jobID == uuid.Nil {
		return nil, status.Error(codes.InvalidArgument, "job_id not valid UUID")
	}

	// set payload
	payload := pollWarehouseTags{
		JobID: jobID,
	}
	middleware.SetTagsPayload(ctx, payload)

	job, err := hub.Engine.Warehouse.PollWarehouseJob(ctx, jobID)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "error polling job: %s", err.Error())
	}

	if job.GetError() != nil {
		payload.Error = job.GetError().Error()
	}
	payload.Status = job.GetStatus().String()
	payload.BytesScanned = job.GetBytesScanned()
	middleware.SetTagsPayload(ctx, payload)

	return &pb.PollWarehouseJobResponse{
		Job: makeWarehouseJobPB(job),
	}, nil
}

func makeWarehouseJobPB(job driver.WarehouseJob) *pb.WarehouseJob {
	var status pb.WarehouseJob_Status
	switch job.GetStatus() {
	case driver.PendingWarehouseJobStatus:
		status = pb.WarehouseJob_PENDING
	case driver.RunningWarehouseJobStatus:
		status = pb.WarehouseJob_RUNNING
	case driver.DoneWarehouseJobStatus:
		status = pb.WarehouseJob_DONE
	default:
		panic(fmt.Errorf("bad job status: %d", job.GetStatus()))
	}

	referencedInstanceIDs := make([][]byte, len(job.GetReferencedInstances()))
	for idx, instanceID := range job.GetReferencedInstances() {
		referencedInstanceIDs[idx] = instanceID.GetStreamInstanceID().Bytes()
	}

	errStr := ""
	if job.GetError() != nil {
		errStr = job.GetError().Error()
	}

	var cursors [][]byte
	for _, engineCursor := range job.GetReplayCursors() {
		cursors = append(cursors, wrapCursor(util.WarehouseCursorType, job.GetJobID(), engineCursor))
	}

	return &pb.WarehouseJob{
		JobId:                 job.GetJobID().Bytes(),
		Status:                status,
		Error:                 errStr,
		ResultAvroSchema:      job.GetResultAvroSchema(),
		ReplayCursors:         cursors,
		ReferencedInstanceIds: referencedInstanceIDs,
		BytesScanned:          job.GetBytesScanned(),
		ResultSizeBytes:       job.GetResultSizeBytes(),
		ResultSizeRecords:     job.GetResultSizeRecords(),
	}
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
