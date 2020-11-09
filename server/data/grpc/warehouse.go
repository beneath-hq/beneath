package grpc

import (
	"context"
	"fmt"

	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"gitlab.com/beneath-hq/beneath/services/data"
	pb "gitlab.com/beneath-hq/beneath/server/data/grpc/proto"
	"gitlab.com/beneath-hq/beneath/infra/engine/driver"
)

func (s *gRPCServer) QueryWarehouse(ctx context.Context, req *pb.QueryWarehouseRequest) (*pb.QueryWarehouseResponse, error) {
	res, err := s.Service.HandleQueryWarehouse(ctx, &data.QueryWarehouseRequest{
		Query:           req.Query,
		Partitions:      req.Partitions,
		DryRun:          req.DryRun,
		TimeoutMs:       req.TimeoutMs,
		MaxBytesScanned: req.MaxBytesScanned,
	})
	if err != nil {
		return nil, err.GRPC()
	}

	return &pb.QueryWarehouseResponse{
		Job: makeWarehouseJobPB(res.Job),
	}, nil
}

func (s *gRPCServer) PollWarehouseJob(ctx context.Context, req *pb.PollWarehouseJobRequest) (*pb.PollWarehouseJobResponse, error) {
	jobID := uuid.FromBytesOrNil(req.JobId)
	if jobID == uuid.Nil {
		return nil, status.Error(codes.InvalidArgument, "job_id not valid UUID")
	}

	res, err := s.Service.HandlePollWarehouse(ctx, &data.PollWarehouseRequest{
		JobID: jobID,
	})
	if err != nil {
		return nil, err.GRPC()
	}

	return &pb.PollWarehouseJobResponse{
		Job: makeWarehouseJobPB(res.Job),
	}, nil
}

func makeWarehouseJobPB(job *data.WarehouseJob) *pb.WarehouseJob {
	var status pb.WarehouseJob_Status
	switch job.Status {
	case driver.PendingWarehouseJobStatus:
		status = pb.WarehouseJob_PENDING
	case driver.RunningWarehouseJobStatus:
		status = pb.WarehouseJob_RUNNING
	case driver.DoneWarehouseJobStatus:
		status = pb.WarehouseJob_DONE
	default:
		panic(fmt.Errorf("bad job status: %d", job.Status))
	}

	referencedInstanceIDs := make([][]byte, len(job.ReferencedInstances))
	for idx, instanceID := range job.ReferencedInstances {
		referencedInstanceIDs[idx] = instanceID.Bytes()
	}

	errStr := ""
	if job.Error != nil {
		errStr = job.Error.Error()
	}

	return &pb.WarehouseJob{
		JobId:                 job.JobID.Bytes(),
		Status:                status,
		Error:                 errStr,
		ResultAvroSchema:      job.ResultAvroSchema,
		ReplayCursors:         job.ReplayCursors,
		ReferencedInstanceIds: referencedInstanceIDs,
		BytesScanned:          job.BytesScanned,
		ResultSizeBytes:       job.ResultSizeBytes,
		ResultSizeRecords:     job.ResultSizeRecords,
	}
}
