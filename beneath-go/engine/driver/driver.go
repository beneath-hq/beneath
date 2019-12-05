package driver

import (
	"context"
	"time"

	"github.com/xtgo/uuid"
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

	// ReadRecords should return up to limit records starting at offset
	ReadRecords(ctx context.Context, p Project, s Stream, i StreamInstance, offset int, limit int) (RecordsReader, error)

	// AppendRecords should insert the records in r at the next free offset
	AppendRecords(ctx context.Context, p Project, s Stream, i StreamInstance, r RecordsReader) (offset int, err error)
}

// LookupService encapsulates functionality to efficiently lookup indexed records in an instance
type LookupService interface {
	Service

	// WriteRecords should insert the records in r for lookup in the given instance
	WriteRecords(ctx context.Context, p Project, s Stream, i StreamInstance, r RecordsReader) error
}

// WarehouseService encapsulates functionality to analytically query records in an instance
type WarehouseService interface {
	Service

	// WriteRecords should insert the records in r for querying on the given instance
	WriteRecords(ctx context.Context, p Project, s Stream, i StreamInstance, r RecordsReader) error
}

// Project encapsulates metadata about a Beneath project
type Project interface {
	GetProjectID() uuid.UUID
	GetName() string
	GetDisplayName() string
	GetDescription() string
	GetPublic() bool
}

// Stream encapsulates metadata about a Beneath stream
type Stream interface {
	GetStreamID() uuid.UUID
	GetName() string
	GetDescription() string
	GetRetention() time.Duration
	GetAvroSchema() string
	GetKeyFields() []string
	EncodeAvro(structured []map[string]interface{}) ([]byte, error)
	DecodeAvro(avro []byte) ([]map[string]interface{}, error)
}

// StreamInstance encapsulates metadata about a Beneath stream instance
type StreamInstance interface {
	GetStreamInstanceID() uuid.UUID
}

// RecordsReader allows iterating over a list of records in various formats
type RecordsReader interface {
	Count() int
	GetKey(idx int) []byte // maybe remove if add cdec to Stream?
	GetTimestamp(idx int) time.Time
	GetAvro(idx int) []byte
	GetStructured(idx int) []map[string]interface{}
}