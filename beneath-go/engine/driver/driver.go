package driver

import (
	"context"
	"time"

	"github.com/beneath-core/beneath-go/core/codec"
	"github.com/beneath-core/beneath-go/core/queryparse"

	uuid "github.com/satori/go.uuid"
)

// MessageQueue encapsulates functionality necessary for message passing in Beneath
type MessageQueue interface {
	// MaxMessageSize should return the maximum allowed byte size of published messages
	MaxMessageSize() int

	// RegisterTopic should register a new topic for message passing
	RegisterTopic(name string) error

	// Publish should issue a new message to all the topic's subscribers
	Publish(topic string, msg []byte) error

	// Subscribe should create a subscription for new messages on the topic.
	// If persistant, messages missed when offline should accumulate and be delivered on reconnect.
	Subscribe(topic string, name string, persistant bool, fn func(ctx context.Context, msg []byte) error) error
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
}

// Log encapsulates functionality necessary to replay instance data in Beneath
type Log interface {
	Service

	// ParseLogQuery should return a cursor for the given log query
	ParseLogQuery(ctx context.Context, p Project, s Stream, i StreamInstance, where queryparse.Query) ([]byte, error)

	// ReadLog should return a RecordsIterator returning records from the log starting at the given cursor
	ReadLog(ctx context.Context, p Project, s Stream, i StreamInstance, cursor []byte, limit int) (RecordsIterator, error)

	// WriteToLog should insert the records in r into the instance's log
	WriteToLog(ctx context.Context, p Project, s Stream, i StreamInstance, r RecordsIterator) error
}

// LookupService encapsulates functionality to efficiently lookup indexed records in an instance
type LookupService interface {
	Service

	// ParseLookupQuery should return a cursor for the given lookup query
	ParseLookupQuery(ctx context.Context, p Project, s Stream, i StreamInstance, where queryparse.Query) ([]byte, error)

	// ReadLookup should return a RecordsIterator looking up records starting at the given cursor
	ReadLookup(ctx context.Context, p Project, s Stream, i StreamInstance, cursor []byte, limit int) (RecordsIterator, error)

	// WriteRecords should insert the records in r for lookup in the instance
	WriteToLookup(ctx context.Context, p Project, s Stream, i StreamInstance, r RecordsIterator) error
}

// WarehouseService encapsulates functionality to analytically query records in an instance
type WarehouseService interface {
	Service

	// WriteToWarehouse should insert the records in r for querying on the given instance
	WriteToWarehouse(ctx context.Context, p Project, s Stream, i StreamInstance, r RecordsIterator) error
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
}

// StreamInstance encapsulates metadata about a Beneath stream instance
type StreamInstance interface {
	// GetStreamInstanceID should return the instance's ID
	GetStreamInstanceID() uuid.UUID
}

// RecordsIterator allows iterating over a list of records in various formats
type RecordsIterator interface {
	// Next should return the next record in the iterator or nil if it's emptied
	Next() Record

	// Len should return the number of rows in the iterator if possible, else -1
	Len() int
}

// Record todo
type Record interface {
	// GetTimestamp should return the timestamp associated with the record
	GetTimestamp() time.Time

	// GetAvro should return the Avro representation of the record
	GetAvro() []byte

	// GetStructured should return the structured representation of the record
	GetStructured() map[string]interface{}
}
