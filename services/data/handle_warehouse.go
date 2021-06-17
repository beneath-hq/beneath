package data

import (
	"context"
	"net/http"

	uuid "github.com/satori/go.uuid"

	"github.com/beneath-hq/beneath/infra/engine/driver"
	"github.com/beneath-hq/beneath/models"
	"github.com/beneath-hq/beneath/services/middleware"
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
func (s *Service) HandleQueryWarehouse(ctx context.Context, req *QueryWarehouseRequest) (*QueryWarehouseResponse, *Error) {
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
	query, err := s.Engine.ExpandWarehouseQuery(ctx, req.Query, func(ctx context.Context, organizationName, projectName, tableName string) (driver.Project, driver.Table, driver.TableInstance, error) {
		// get instance ID
		instanceID := s.Tables.FindPrimaryInstanceIDByOrganizationProjectAndName(ctx, organizationName, projectName, tableName)
		if instanceID == uuid.Nil {
			return nil, nil, nil, newErrorf(http.StatusNotFound, "instance not found for table '%s/%s/%s'", organizationName, projectName, tableName)
		}

		// get table
		table := s.Tables.FindCachedInstance(ctx, instanceID)
		if table == nil {
			return nil, nil, nil, newErrorf(http.StatusNotFound, "table '%s/%s/%s' not found", organizationName, projectName, tableName)
		}

		// check permissions
		perms := s.Permissions.TablePermissionsForSecret(ctx, secret, table.TableID, table.ProjectID, table.Public)
		if !perms.Read {
			return nil, nil, nil, newErrorf(http.StatusForbidden, "token doesn't grant right to read table '%s/%s/%s'", organizationName, projectName, tableName)
		}

		// success
		return table, table, models.EfficientTableInstance(instanceID), nil
	})
	if err != nil {
		if errr, ok := err.(*Error); ok {
			return nil, errr
		}
		return nil, newError(http.StatusBadRequest, err.Error())
	}

	// analyze query
	analyzeJob, err := s.Engine.Warehouse.AnalyzeWarehouseQuery(ctx, query)
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
	err = s.Usage.CheckScanQuota(ctx, secret, estimatedBytesScanned)
	if err != nil {
		return nil, newError(http.StatusTooManyRequests, err.Error())
	}

	// run real job
	jobID := uuid.NewV4()
	job, err := s.Engine.Warehouse.RunWarehouseQuery(ctx, jobID, query, int(req.Partitions), int(req.TimeoutMs), int(req.MaxBytesScanned))
	if err != nil {
		return nil, newErrorf(http.StatusBadRequest, "error running query: %s", err.Error())
	}

	// assign analyze fields not returned from run (kind of a hack..., fix this when we persist job objects)

	// update payload
	payload.JobID = &jobID
	middleware.SetTagsPayload(ctx, payload)

	// track read usage
	s.Usage.TrackScan(ctx, secret, estimatedBytesScanned)

	return &QueryWarehouseResponse{Job: wrapWarehouseJob(job, analyzeJob)}, nil
}

// HandlePollWarehouse handles a warehouse poll request
func (s *Service) HandlePollWarehouse(ctx context.Context, req *PollWarehouseRequest) (*PollWarehouseResponse, *Error) {
	// set payload
	payload := pollWarehouseTags{
		JobID: req.JobID,
	}
	middleware.SetTagsPayload(ctx, payload)

	job, err := s.Engine.Warehouse.PollWarehouseJob(ctx, req.JobID)
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
		cursors = append(cursors, wrapCursor(WarehouseCursorType, job.GetJobID(), engineCursor))
	}

	// coalesce GetReferencedInstances with analyzeJob
	referencedInstances := job.GetReferencedInstances()
	if analyzeJob != nil && len(referencedInstances) == 0 {
		referencedInstances = analyzeJob.GetReferencedInstances()
	}

	referencedInstanceIDs := make([]uuid.UUID, len(referencedInstances))
	for idx, instanceID := range referencedInstances {
		referencedInstanceIDs[idx] = instanceID.GetTableInstanceID()
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
