package engine

import (
	"context"
	"time"

	"github.com/beneath-core/beneath-go/core/codec"
	pb "github.com/beneath-core/beneath-go/proto"
	uuid "github.com/satori/go.uuid"
)

// StreamsDriver defines the functions necessary to encapsulate Beneath's streaming data needs
type StreamsDriver interface {
	// GetMaxMessageSize returns the maximum accepted message size in bytes
	GetMaxMessageSize() int

	// QueueWriteRequest queues a write request -- concretely, it results in
	// the write request being written to Pubsub, then from there read by
	// the data processing pipeline and written to BigTable and BigQuery
	QueueWriteRequest(ctx context.Context, req *pb.WriteRecordsRequest) error

	// ReadWriteRequests triggers fn for every WriteRecordsRequest that's written with QueueWriteRequest
	ReadWriteRequests(fn func(context.Context, *pb.WriteRecordsRequest) error) error

	// QueueWriteReport publishes a batch of keys + metrics to the streams driver
	QueueWriteReport(ctx context.Context, rep *pb.WriteRecordsReport) error

	// ReadWriteReports reads messages from the Metrics topic
	ReadWriteReports(fn func(context.Context, *pb.WriteRecordsReport) error) error

	// QueueTask queues a task for processing
	QueueTask(ctx context.Context, t *pb.QueuedTask) error

	// ReadTasks reads queued tasks
	ReadTasks(fn func(context.Context, *pb.QueuedTask) error) error
}

// TablesDriver defines the functions necessary to encapsulate Beneath's operational datastore needs
type TablesDriver interface {
	// GetMaxKeySize returns the maximum accepted key size in bytes
	GetMaxKeySize() int

	// GetMaxDataSize returns the maximum accepted value size in bytes
	GetMaxDataSize() int

	// GetMaxBatchLength returns the maximum accepted number of rows per batch
	GetMaxBatchLength() int

	// WriteRecords saves one or multiple records. It does not save records if timestamp is lower than that of a previous write to the same key
	WriteRecords(ctx context.Context, instanceID uuid.UUID, keys [][]byte, avroData [][]byte, timestamps []time.Time, saveLatest bool) error

	// ReadRecords reads one or multiple (not necessarily sequential) records by key and calls fn one by one
	ReadRecords(ctx context.Context, instanceID uuid.UUID, keys [][]byte, fn func(idx uint, avroData []byte, timestamp time.Time) error) error

	// ReadRecordRange reads one or a range of records by key and calls fn one by one
	ReadRecordRange(ctx context.Context, instanceID uuid.UUID, keyRange codec.KeyRange, limit int, fn func(avroData []byte, timestamp time.Time) error) error

	// ReadLatestRecords returns the latest records written to the instance
	ReadLatestRecords(ctx context.Context, instanceID uuid.UUID, limit int, before time.Time, fn func(avroData []byte, timestamp time.Time) error) error

	// ClearRecords removes all records for an instance
	ClearRecords(ctx context.Context, instanceID uuid.UUID) error

	// CommitUsage writes a batch of usage metrics
	CommitUsage(ctx context.Context, key []byte, usage pb.QuotaUsage) error

	// ReadSingleUsage reads usage metrics for one key
	ReadSingleUsage(ctx context.Context, key []byte) (pb.QuotaUsage, error)

	// ReadUsage reads usage metrics for multiple periods and calls fn one by one
	ReadUsage(ctx context.Context, fromKey []byte, toKey []byte, fn func(key []byte, usage pb.QuotaUsage) error) error
}

// WarehouseDriver defines the functions necessary to encapsulate Beneath's data archiving needs
type WarehouseDriver interface {
	// GetMaxDataSize returns the maximum accepted row size in bytes
	GetMaxDataSize() int

	// GetMaxBatchLength returns the maximum accepted number of rows per batch
	GetMaxBatchLength() int

	// WriteRecords saves one or multiple records to the data warehouse
	WriteRecords(ctx context.Context, projectName string, streamName string, instanceID uuid.UUID, keys [][]byte, avros [][]byte, records []map[string]interface{}, timestamps []time.Time) error

	// RegisterProject should be called when a new project is created to create a corresponding dataset in the warehouse
	RegisterProject(ctx context.Context, projectID uuid.UUID, public bool, name, displayName, description string) error

	// UpdateProject should be called to update metadata on a project
	UpdateProject(ctx context.Context, projectID uuid.UUID, public bool, name, displayName, description string) error

	// DeregisterProject should be called when a project is deleted
	DeregisterProject(ctx context.Context, projectID uuid.UUID, name string) error

	// RegisterStreamInstance should be called when a new stream instance is created to create a corresponding table in the warehouse
	RegisterStreamInstance(ctx context.Context, projectName string, streamID uuid.UUID, streamName string, streamDescription string, schemaJSON string, keyFields []string, instanceID uuid.UUID) error

	// PromoteStreamInstance should be called to promote an instance to the current data source for the stream
	PromoteStreamInstance(ctx context.Context, projectName string, streamID uuid.UUID, streamName string, streamDescription string, instanceID uuid.UUID) error

	// UpdateStreamInstance should be called to update metadata on a stream instance
	UpdateStreamInstance(ctx context.Context, projectName string, streamName string, streamDescription string, schemaJSON string, instanceID uuid.UUID) error

	// DeregisterStreamInstance should be called to remove a stream instance
	DeregisterStreamInstance(ctx context.Context, projectName string, streamName string, instanceID uuid.UUID) error
}