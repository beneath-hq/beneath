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
	Query           string
	DryRun          bool
	TimeoutMs       int32
	MaxBytesScanned int64
}

type warehouseJob struct {
	JobID               *uuid.UUID  `json:"job_id,omitempty"`
	Status              string      `json:"status,omitempty"`
	Error               string      `json:"error,omitempty"`
	ResultAvroSchema    string      `json:"result_avro_schema,omitempty"`
	ReplayCursors       []string    `json:"replay_cursors,omitempty"`
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

	args.Query = r.URL.Query().Get("query")

	dry, err := parseBoolParam("dry", r.URL.Query().Get("dry"))
	if err != nil {
		return args, httputil.NewError(http.StatusBadRequest, err.Error())
	}
	args.DryRun = dry

	timeout, err := parseIntParam("timeout_ms", r.URL.Query().Get("timeout_ms"))
	if err != nil {
		return args, httputil.NewError(http.StatusBadRequest, err.Error())
	}
	args.TimeoutMs = int32(timeout)

	maxBytesScanned, err := parseIntParam("max_bytes_scanned", r.URL.Query().Get("max_bytes_scanned"))
	if err != nil {
		return args, httputil.NewError(http.StatusBadRequest, err.Error())
	}
	args.MaxBytesScanned = int64(maxBytesScanned)

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

	encode.ReplayCursors = make([]string, len(job.ReplayCursors))
	for i, cursor := range job.ReplayCursors {
		encode.ReplayCursors[i] = base58.Encode(cursor)
	}

	w.Header().Set("Content-Type", "application/json")
	err := jsonutil.MarshalWriter(encode, w)
	if err != nil {
		return err
	}

	return nil
}
