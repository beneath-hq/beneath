package api

import (
	"context"
	"net/http"

	uuid "github.com/satori/go.uuid"

	"gitlab.com/beneath-hq/beneath/control/entity"
	"gitlab.com/beneath-hq/beneath/engine/driver"
	"gitlab.com/beneath-hq/beneath/gateway/util"
	"gitlab.com/beneath-hq/beneath/hub"
	"gitlab.com/beneath-hq/beneath/internal/middleware"
)

// QueryWarehouseRequest is a request to HandleQueryWarehouse
type QueryWarehouseRequest struct {
	Query           string
	Partitions      int32
	DryRun          bool
	TimeoutMs       int32
	MaxBytesScanned int64
}

// QueryWarehouseResponse is a result from HandleQueryWarehouse
type QueryWarehouseResponse struct {
	Job *WarehouseJob
}

// PollWarehouseRequest is a request to HandlePollWarehouse
type PollWarehouseRequest struct {
	JobID uuid.UUID
}

// PollWarehouseResponse is a result from HandlePollWarehouse
type PollWarehouseResponse struct {
	Job *WarehouseJob
}

// WarehouseJob represents a warehouse query job
type WarehouseJob struct {
	JobID               uuid.UUID
	Status              driver.WarehouseJobStatus
	Error               error
	ResultAvroSchema    string
	ReplayCursors       [][]byte
	ReferencedInstances []uuid.UUID
	BytesScanned        int64
	ResultSizeBytes     int64
	ResultSizeRecords   int64
}

type queryWarehouseTags struct {
	JobID           *uuid.UUID `json:"job,omitempty"`
	Query           string     `json:"query,omitempty"`
	Partitions      int32      `json:"partitions,omitempty"`
	DryRun          bool       `json:"dry,omitempty"`
	TimeoutMs       int32      `json:"timeout,omitempty"`
	MaxBytesScanned int64      `json:"max_scan,omitempty"`
}

type pollWarehouseTags struct {
	JobID        uuid.UUID `json:"job,omitempty"`
	Status       string    `json:"status,omitempty"`
	Error        string    `json:"error,omitempty"`
	BytesScanned int64     `json:"scan,omitempty"`
}

// HandleQueryWarehouse handles a warehouse query request
func HandleQueryWarehouse(ctx context.Context, req *QueryWarehouseRequest) (*QueryWarehouseResponse, *Error) {
	// get auth
	secret := middleware.GetSecret(ctx)
	if secret == nil || secret.IsAnonymous() {
		return nil, newErrorf(http.StatusUnauthorized, "not authenticated")
	}

	// set payload
	payload := queryWarehouseTags{
		Query:           req.Query,
		Partitions:      req.Partitions,
		DryRun:          req.DryRun,
		TimeoutMs:       req.TimeoutMs,
		MaxBytesScanned: req.MaxBytesScanned,
	}
	middleware.SetTagsPayload(ctx, payload)

	// checks
	if req.Query == "" {
		return nil, newError(http.StatusBadRequest, "parameter 'query' is empty")
	}

	// expand query
	query, err := hub.Engine.ExpandWarehouseQuery(ctx, req.Query, func(ctx context.Context, organizationName, projectName, streamName string) (driver.Project, driver.Stream, driver.StreamInstance, error) {
		// get instance ID
		instanceID := entity.FindInstanceIDByOrganizationProjectAndName(ctx, organizationName, projectName, streamName)
		if instanceID == uuid.Nil {
			return nil, nil, nil, newErrorf(http.StatusNotFound, "instance not found for stream '%s/%s/%s'", organizationName, projectName, streamName)
		}

		// get stream
		stream := entity.FindCachedStreamByCurrentInstanceID(ctx, instanceID)
		if stream == nil {
			return nil, nil, nil, newErrorf(http.StatusNotFound, "stream '%s/%s/%s' not found", organizationName, projectName, streamName)
		}

		// check permissions
		perms := secret.StreamPermissions(ctx, stream.StreamID, stream.ProjectID, stream.Public)
		if !perms.Read {
			return nil, nil, nil, newErrorf(http.StatusForbidden, "token doesn't grant right to read stream '%s/%s/%s'", organizationName, projectName, streamName)
		}

		// success
		return stream, stream, entity.EfficientStreamInstance(instanceID), nil
	})
	if err != nil {
		if errr, ok := err.(*Error); ok {
			return nil, errr
		}
		return nil, newError(http.StatusBadRequest, err.Error())
	}

	// analyze query
	analyzeJob, err := hub.Engine.Warehouse.AnalyzeWarehouseQuery(ctx, query)
	if err != nil {
		return nil, newErrorf(http.StatusBadRequest, "error during query analysis: %s", err.Error())
	}

	// stop if dry
	if req.DryRun {
		return &QueryWarehouseResponse{Job: wrapWarehouseJob(analyzeJob, nil)}, nil
	}

	// check bytes scanned
	estimatedBytesScanned := analyzeJob.GetBytesScanned()
	if req.MaxBytesScanned != 0 && estimatedBytesScanned > req.MaxBytesScanned {
		return nil, newErrorf(http.StatusTooManyRequests, "query would scan %d bytes, which exceeds job limit of %d (limit the query or increase the limit)", analyzeJob.GetBytesScanned(), req.MaxBytesScanned)
	}

	// check quota
	err = util.CheckScanQuota(ctx, secret, estimatedBytesScanned)
	if err != nil {
		return nil, newError(http.StatusTooManyRequests, err.Error())
	}

	// run real job
	jobID := uuid.NewV4()
	job, err := hub.Engine.Warehouse.RunWarehouseQuery(ctx, jobID, query, int(req.Partitions), int(req.TimeoutMs), int(req.MaxBytesScanned))
	if err != nil {
		return nil, newErrorf(http.StatusBadRequest, "error running query: %s", err.Error())
	}

	// assign analyze fields not returned from run (kind of a hack..., fix this when we persist job objects)

	// update payload
	payload.JobID = &jobID
	middleware.SetTagsPayload(ctx, payload)

	// track read metrics
	util.TrackScan(ctx, secret, estimatedBytesScanned)

	return &QueryWarehouseResponse{Job: wrapWarehouseJob(job, analyzeJob)}, nil
}

// HandlePollWarehouse handles a warehouse poll request
func HandlePollWarehouse(ctx context.Context, req *PollWarehouseRequest) (*PollWarehouseResponse, *Error) {
	// set payload
	payload := pollWarehouseTags{
		JobID: req.JobID,
	}
	middleware.SetTagsPayload(ctx, payload)

	job, err := hub.Engine.Warehouse.PollWarehouseJob(ctx, req.JobID)
	if err != nil {
		return nil, newErrorf(http.StatusBadRequest, "error polling job: %s", err.Error())
	}

	if job.GetError() != nil {
		payload.Error = job.GetError().Error()
	}
	payload.Status = job.GetStatus().String()
	payload.BytesScanned = job.GetBytesScanned()
	middleware.SetTagsPayload(ctx, payload)

	return &PollWarehouseResponse{Job: wrapWarehouseJob(job, nil)}, nil
}

func wrapWarehouseJob(job driver.WarehouseJob, analyzeJob driver.WarehouseJob) *WarehouseJob {
	var cursors [][]byte
	for _, engineCursor := range job.GetReplayCursors() {
		cursors = append(cursors, wrapCursor(util.WarehouseCursorType, job.GetJobID(), engineCursor))
	}

	// coalesce GetReferencedInstances with analyzeJob
	referencedInstances := job.GetReferencedInstances()
	if analyzeJob != nil && len(referencedInstances) == 0 {
		referencedInstances = analyzeJob.GetReferencedInstances()
	}

	referencedInstanceIDs := make([]uuid.UUID, len(referencedInstances))
	for idx, instanceID := range referencedInstances {
		referencedInstanceIDs[idx] = instanceID.GetStreamInstanceID()
	}

	res := &WarehouseJob{
		JobID:               job.GetJobID(),
		Status:              job.GetStatus(),
		Error:               job.GetError(),
		ResultAvroSchema:    job.GetResultAvroSchema(),
		ReplayCursors:       cursors,
		ReferencedInstances: referencedInstanceIDs,
		BytesScanned:        job.GetBytesScanned(),
		ResultSizeBytes:     job.GetResultSizeBytes(),
		ResultSizeRecords:   job.GetResultSizeRecords(),
	}

	// coalesce GetResultAvroSchema with analyzeJob
	if analyzeJob != nil && res.ResultAvroSchema == "" {
		res.ResultAvroSchema = analyzeJob.GetResultAvroSchema()
	}

	// coalesce GetBytesScanned with analyzeJob
	if analyzeJob != nil && res.BytesScanned == 0 {
		res.BytesScanned = analyzeJob.GetBytesScanned()
	}

	return res
}
