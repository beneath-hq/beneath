package driver

import (
	"context"
	"time"

	"gitlab.com/beneath-hq/beneath/pkg/timeutil"

	uuid "github.com/satori/go.uuid"

	pb "gitlab.com/beneath-hq/beneath/engine/proto"
	"gitlab.com/beneath-hq/beneath/pkg/codec"
	"gitlab.com/beneath-hq/beneath/pkg/queryparse"
)

// MessageQueue encapsulates functionality necessary for message passing in Beneath
type MessageQueue interface {
	// MaxMessageSize should return the maximum allowed byte size of published messages
	MaxMessageSize() int

	// RegisterTopic should register a new topic for message passing
	RegisterTopic(name string) error

	// Publish should issue a new message to all the topic's subscribers
	Publish(ctx context.Context, topic string, msg []byte) error

	// Subscribe should create a subscription for new messages on the topic.
	// If persistent, messages missed when offline should accumulate and be delivered on reconnect.
	Subscribe(ctx context.Context, topic string, name string, persistent bool, fn func(ctx context.Context, msg []byte) error) error

	// Reset should clear all data in the service (useful during testing)
	Reset(ctx context.Context) error
}

// Service encapsulates functionality expected of components that store instance data in Beneath
type Service interface {
	// MaxKeySize should return the maximum allowed byte size of a single record's key
	MaxKeySize() int

	// MaxRecordSize should return the maximum allowed byte size of a single record
	MaxRecordSize() int

	// MaxRecordsInBatch should return the highest number of records that may be passed in one RecordsReader
	MaxRecordsInBatch() int

	// RegisterProject is called when a project is created *or updated*
	RegisterProject(ctx context.Context, p Project) error

	// RemoveProject is called when a project is deleted
	RemoveProject(ctx context.Context, p Project) error

	// RegisterInstance is called when a new instance is created
	RegisterInstance(ctx context.Context, p Project, s Stream, i StreamInstance) error

	// PromoteInstance is called when an instance is promoted to be the main instance of the stream
	PromoteInstance(ctx context.Context, p Project, s Stream, i StreamInstance) error

	// RemoveInstance is called when an instance is deleted
	RemoveInstance(ctx context.Context, p Project, s Stream, i StreamInstance) error

	// Reset should clear all data in the service (useful during testing)
	Reset(ctx context.Context) error
}

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

	// NOTE: Usage tracking is temporarily implemented here, but really should be implemented on top of the engine, not in it

	// CommitUsage writes a batch of usage metrics
	CommitUsage(ctx context.Context, id uuid.UUID, period timeutil.Period, ts time.Time, usage pb.QuotaUsage) error

	// ReadSingleUsage reads usage metrics for one key
	ReadSingleUsage(ctx context.Context, id uuid.UUID, period timeutil.Period, ts time.Time) (pb.QuotaUsage, error)

	// ReadUsage reads usage metrics for multiple periods and calls fn one by one
	ReadUsage(ctx context.Context, id uuid.UUID, period timeutil.Period, from time.Time, until time.Time, fn func(ts time.Time, usage pb.QuotaUsage) error) error

	// ClearUsage clears all usage data saved for the id
	ClearUsage(ctx context.Context, id uuid.UUID) error
}

// WarehouseService encapsulates functionality to analytically query records in an instance
type WarehouseService interface {
	Service

	// WriteToWarehouse should insert the records in r for querying on the given instance
	WriteToWarehouse(ctx context.Context, p Project, s Stream, i StreamInstance, rs []Record) error
}

// Project encapsulates metadata about a Beneath project
type Project interface {
	// GetProjectID should return the project ID
	GetProjectID() uuid.UUID

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

	// GetRetention should return the amount of time data should be retained, or 0 for infinite retention
	GetRetention() time.Duration

	// GetCodec should return a codec for serializing and deserializing stream data and keys
	GetCodec() *codec.Codec

	// GetBigQuerySchema should return the stream's schema in BigQuery JSON representation
	// TODO: Remove by creating an avro schema -> bigquery schema transpiler
	GetBigQuerySchema() string
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

	// GetPrimaryKey should return a byte-encoded representation of the record's primary key
	// Only implemented on driver output, not necessary for input (where key is computed from schema and data)
	// TODO: not beautiful/necessary; it's a convenience to avoid recomputing the primary key in the REST gateway
	GetPrimaryKey() []byte
}
