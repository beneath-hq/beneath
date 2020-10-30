package driver

import (
	"context"
	"fmt"
	"time"

	uuid "github.com/satori/go.uuid"

	pb "gitlab.com/beneath-hq/beneath/infrastructure/engine/proto"
	"gitlab.com/beneath-hq/beneath/pkg/codec"
	"gitlab.com/beneath-hq/beneath/pkg/queryparse"
)

// Tracking drivers
// ----------------

// Driver is a function that creates an engine service (LookupService or WarehouseService) from a config object
type Driver func(opts map[string]interface{}) (Service, error)

// Drivers is a registry of drivers
var Drivers = make(map[string]Driver)

// AddDriver registers a new driver (by passing the driver's constructor)
func AddDriver(name string, driver Driver) {
	if Drivers[name] != nil {
		panic(fmt.Errorf("Service driver already registered with name '%s'", name))
	}
	Drivers[name] = driver
}

// Base
// ------

// Project encapsulates metadata about a Beneath project
type Project interface {
	// GetProjectID should return the project ID
	GetProjectID() uuid.UUID

	// GetOrganizationName should return the project's organization's name
	GetOrganizationName() string

	// GetProjectName should return the project's identifying name (not its display name)
	GetProjectName() string

	// GetPublic should return true if the project is publicly accessible
	GetPublic() bool
}

// Stream encapsulates metadata about a Beneath stream
type Stream interface {
	// GetStreamID should return the stream ID
	GetStreamID() uuid.UUID

	// GetStreamName should return the stream's identifying name
	GetStreamName() string

	// GetUseLog should return true if records should be stored for log-based replay and change capture
	GetUseLog() bool

	// GetUseIndex should return true if records should be stored for indexed lookup
	GetUseIndex() bool

	// GetUseWarehouse should return true if records should be saved for data warehouse queries
	GetUseWarehouse() bool

	// GetLogRetention should return the duration data should be retained (use 0 for infinite retention)
	GetLogRetention() time.Duration

	// GetIndexRetention should return the duration data should be retained (use 0 for infinite retention)
	GetIndexRetention() time.Duration

	// GetWarehouseRetention should return the duration data should be retained (use 0 for infinite retention)
	GetWarehouseRetention() time.Duration

	// GetCodec should return a codec for serializing and deserializing stream data and keys
	GetCodec() *codec.Codec
}

// StreamInstance encapsulates metadata about a Beneath stream instance
type StreamInstance interface {
	// GetStreamInstanceID should return the instance's ID
	GetStreamInstanceID() uuid.UUID
}

// RecordsIterator allows iterating over a list of records in various formats
type RecordsIterator interface {
	// Next should advance the iterator and return true if succesful (i.e. there's another record)
	Next() bool

	// Record should return the current record. If Next hasn't been called yet or the latest call to Next returned false, the result is undefined.
	Record() Record

	// NextCursor should return a cursor for reading more rows if this iterator
	// doesn't return all rows in the result. A nil result should indicate no more rows.
	NextCursor() []byte
}

// Record todo
type Record interface {
	// GetTimestamp should return the timestamp associated with the record
	GetTimestamp() time.Time

	// GetAvro should return the Avro representation of the record
	GetAvro() []byte

	// GetStructured should return the structured representation of the record
	GetStructured() map[string]interface{}

	// GetJSON should return a JSON-serializable representation of the record
	GetJSON() map[string]interface{}

	// GetPrimaryKey should return a byte-encoded representation of the record's primary key
	// Only implemented on driver output, not necessary for input (where key is computed from schema and data)
	// TODO: not beautiful/necessary; it's a convenience to avoid recomputing the primary key in the REST gateway
	GetPrimaryKey() []byte
}

// Services base
// -------------

// Service encapsulates functionality expected of components that store instance data in Beneath
type Service interface {
	// AsLookupService should return the service cast as a LookupService if supported, else nil
	AsLookupService() LookupService

	// AsWarehouseService should return the service cast as a WarehouseService if supported, else nil
	AsWarehouseService() WarehouseService

	// AsUsageService should return the service cast as a UsageService if supported, else nil
	AsUsageService() UsageService

	// MaxKeySize should return the maximum allowed byte size of a single record's key
	MaxKeySize() int

	// MaxRecordSize should return the maximum allowed byte size of a single record
	MaxRecordSize() int

	// MaxRecordsInBatch should return the highest number of records that may be passed in one RecordsReader
	MaxRecordsInBatch() int

	// RegisterInstance is called when a new instance is created
	RegisterInstance(ctx context.Context, s Stream, i StreamInstance) error

	// RemoveInstance is called when an instance is deleted
	RemoveInstance(ctx context.Context, s Stream, i StreamInstance) error

	// Reset should clear all data in the service (useful during testing)
	Reset(ctx context.Context) error
}

// Lookup
// ------

// LookupService encapsulates functionality necessary to lookup, replay and subscribe to data.
// It's implementation is likely a mixture of a log and an indexed operational data store.
type LookupService interface {
	Service

	// ParseQuery returns (replayCursors, changeCursors, err)
	ParseQuery(ctx context.Context, p Project, s Stream, i StreamInstance, where queryparse.Query, compacted bool, partitions int) ([][]byte, [][]byte, error)

	// Peek returns (rewindCursor, changeCursor, err)
	Peek(ctx context.Context, p Project, s Stream, i StreamInstance) ([]byte, []byte, error)

	// ReadCursor returns (records, nextCursor, err)
	ReadCursor(ctx context.Context, p Project, s Stream, i StreamInstance, cursor []byte, limit int) (RecordsIterator, error)

	// WriteRecord persists the records
	WriteRecords(ctx context.Context, p Project, s Stream, i StreamInstance, rs []Record) error
}

// Warehouse
// ---------

// WarehouseService encapsulates functionality to analytically query records in an instance
type WarehouseService interface {
	Service

	// WriteToWarehouse should insert the records in r for querying on the given instance
	WriteToWarehouse(ctx context.Context, p Project, s Stream, i StreamInstance, rs []Record) error

	// GetWarehouseTableName should return the table name that the driver wants stream references in queries expanded to
	GetWarehouseTableName(p Project, s Stream, i StreamInstance) string

	// AnalyzeWarehouseQuery should perform a dry-run of query
	AnalyzeWarehouseQuery(ctx context.Context, query string) (WarehouseJob, error)

	// RunWarehouseQuery should start a query and return a job. If the job is not completed outright, it should be
	// possible to poll the job with PollWarehouseJob using the job ID.
	RunWarehouseQuery(ctx context.Context, jobID uuid.UUID, query string, partitions int, timeoutMs int, maxBytesScanned int) (WarehouseJob, error)

	// PollWarehouseJob should return the current state of the job identified by jobID
	PollWarehouseJob(ctx context.Context, jobID uuid.UUID) (WarehouseJob, error)

	// ReadWarehouseCursor returns records for the cursor
	ReadWarehouseCursor(ctx context.Context, cursor []byte, limit int) (RecordsIterator, error)
}

// WarehouseJobStatus represents the current status of a warehouse job
type WarehouseJobStatus int

const (
	// UnspecifiedWarehouseJobStatus is the null value for WarehouseJobStatus
	UnspecifiedWarehouseJobStatus WarehouseJobStatus = iota

	// PendingWarehouseJobStatus represents a warehouse job that has not yet been started
	PendingWarehouseJobStatus

	// RunningWarehouseJobStatus represents a warehouse job that is currently running
	RunningWarehouseJobStatus

	// DoneWarehouseJobStatus represents a warehouse job that has completed; if the job failed, the job will have a non-nil error
	DoneWarehouseJobStatus
)

func (s WarehouseJobStatus) String() string {
	switch s {
	case PendingWarehouseJobStatus:
		return "pending"
	case RunningWarehouseJobStatus:
		return "running"
	case DoneWarehouseJobStatus:
		return "done"
	default:
		return "unspecified"
	}
}

// WarehouseJob represents a query job submitted to a WarehouseService
type WarehouseJob interface {
	// GetJobID should return an identifier for the job
	GetJobID() uuid.UUID

	// GetStatus returns the current status of the job
	GetStatus() WarehouseJobStatus

	// GetError returns a non-nil error if the status is Done and the query failed
	GetError() error

	// GetResultAvroSchema returns an Avro schema for the query result. Only set for successful queries.
	GetResultAvroSchema() string

	// GetReplayCursors returns one or more cursors that can be submitted to the WarehouseService's ReadWarehouseCursor to
	// retrieve query results.
	GetReplayCursors() [][]byte

	// GetReferencedInstances returns all the instances referenced in the query
	GetReferencedInstances() []StreamInstance

	// GetBytesScanned returns an estimate of the number of bytes scanned by the query
	GetBytesScanned() int64

	// GetResultSizeBytes returns an estimate of the number of bytes in the query result
	GetResultSizeBytes() int64

	// GetResultSizeRecords returns an estimate of the number of records in the query result
	GetResultSizeRecords() int64
}

// Usage
// -----

// NOTE: Usage tracking is temporarily implemented in the engine, but logically doesn't really belong here.
// It should really be implemented as separate infrastructure (possibly even on top of the engine as "meta" streams).

// UsageLabel defines valid labels for UsageService
type UsageLabel string

// Consts for UsageService
const (
	UsageLabelMonthly    = "month"
	UsageLabelHourly     = "hour"
	UsageLabelQuotaMonth = "quota_month"
)

// UsageService stores and retrieves aggregated usage by timestamp for different resources in Beneath.
// It powers the monitoring dashboards and quota checks.
type UsageService interface {
	// WriteUsage writes a batch of usage metrics
	WriteUsage(ctx context.Context, id uuid.UUID, label UsageLabel, ts time.Time, usage pb.QuotaUsage) error

	// ReadSingleUsage reads usage metrics for one key
	ReadUsageSingle(ctx context.Context, id uuid.UUID, label UsageLabel, ts time.Time) (pb.QuotaUsage, error)

	// ReadUsage reads usage metrics for multiple periods and calls fn one by one
	ReadUsageRange(ctx context.Context, id uuid.UUID, label UsageLabel, from time.Time, until time.Time, limit int, fn func(ts time.Time, usage pb.QuotaUsage) error) error

	// ClearUsage clears all usage data saved for the id
	ClearUsage(ctx context.Context, id uuid.UUID) error

	// Reset should clear all data in the service
	Reset(ctx context.Context) error
}
