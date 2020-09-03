package http

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/mr-tron/base58"
	uuid "github.com/satori/go.uuid"

	"gitlab.com/beneath-hq/beneath/engine/driver"
	"gitlab.com/beneath-hq/beneath/gateway/api"
	"gitlab.com/beneath-hq/beneath/pkg/httputil"
	"gitlab.com/beneath-hq/beneath/pkg/jsonutil"
)

type queryWarehouseArgs struct {
	Query           string `json:"query",omitempty`
	DryRun          bool   `json:"dry",omitempty`
	TimeoutMs       int32  `json:"timeout_ms",omitempty`
	MaxBytesScanned int64  `json:"max_bytes_scanned",omitempty`
}

type warehouseJob struct {
	JobID               *uuid.UUID  `json:"job_id,omitempty"`
	Status              string      `json:"status,omitempty"`
	Error               string      `json:"error,omitempty"`
	ResultAvroSchema    string      `json:"result_avro_schema,omitempty"`
	ReplayCursor        string      `json:"replay_cursor,omitempty"`
	ReferencedInstances []uuid.UUID `json:"referenced_instances,omitempty"`
	BytesScanned        int64       `json:"bytes_scanned,omitempty"`
	ResultSizeBytes     int64       `json:"result_size_bytes,omitempty"`
	ResultSizeRecords   int64       `json:"result_size_records,omitempty"`
}

func postToWarehouseJob(w http.ResponseWriter, r *http.Request) error {
	args, err := parseQueryWarehouseArgs(r)
	if err != nil {
		return err
	}

	res, errr := api.HandleQueryWarehouse(r.Context(), &api.QueryWarehouseRequest{
		Query:           args.Query,
		Partitions:      1,
		DryRun:          args.DryRun,
		TimeoutMs:       args.TimeoutMs,
		MaxBytesScanned: args.MaxBytesScanned,
	})
	if errr != nil {
		return errr.HTTP()
	}

	return handleWarehouseJob(w, res.Job)
}

func getFromWarehouseJob(w http.ResponseWriter, r *http.Request) error {
	jobID, err := uuid.FromString(chi.URLParam(r, "jobID"))
	if err != nil {
		return httputil.NewError(404, "job_id not found or not a valid UUID")
	}

	res, errr := api.HandlePollWarehouse(r.Context(), &api.PollWarehouseRequest{
		JobID: jobID,
	})
	if errr != nil {
		return errr.HTTP()
	}

	return handleWarehouseJob(w, res.Job)
}

func parseQueryWarehouseArgs(r *http.Request) (queryWarehouseArgs, error) {
	args := queryWarehouseArgs{}
	err := jsonutil.Unmarshal(r.Body, &args)
	if err != nil {
		return args, httputil.NewError(http.StatusBadRequest, "request body must be json")
	}

	return args, nil
}

func handleWarehouseJob(w http.ResponseWriter, job *api.WarehouseJob) error {
	encode := warehouseJob{
		ResultAvroSchema:    job.ResultAvroSchema,
		ReferencedInstances: job.ReferencedInstances,
		BytesScanned:        job.BytesScanned,
		ResultSizeBytes:     job.ResultSizeBytes,
		ResultSizeRecords:   job.ResultSizeRecords,
	}

	if job.JobID != uuid.Nil {
		encode.JobID = &job.JobID
	}

	switch job.Status {
	case driver.PendingWarehouseJobStatus:
		encode.Status = "pending"
	case driver.RunningWarehouseJobStatus:
		encode.Status = "running"
	case driver.DoneWarehouseJobStatus:
		encode.Status = "done"
	default:
		panic(fmt.Errorf("bad job status: %d", job.Status))
	}

	if job.Error != nil {
		encode.Error = job.Error.Error()
	}

	if len(job.ReplayCursors) > 0 {
		encode.ReplayCursor = base58.Encode(job.ReplayCursors[0])
	}

	encodeWrapper := map[string]interface{}{
		"data": encode,
	}

	w.Header().Set("Content-Type", "application/json")
	err := jsonutil.MarshalWriter(encodeWrapper, w)
	if err != nil {
		return err
	}

	return nil
}
