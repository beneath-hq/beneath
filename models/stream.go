package models

import (
	"fmt"
	"regexp"
	"time"

	uuid "github.com/satori/go.uuid"
	"gopkg.in/go-playground/validator.v9"

	"gitlab.com/beneath-hq/beneath/pkg/codec"
)

// StreamSchemaKind indicates the SDL of a stream's schema
type StreamSchemaKind string

// Constants for StreamSchemaKind
const (
	StreamSchemaKindGraphQL StreamSchemaKind = "GraphQL"
	StreamSchemaKindAvro    StreamSchemaKind = "Avro"
)

// Stream represents a collection of data
type Stream struct {
	_msgpack struct{} `msgpack:",omitempty"`

	// Descriptive fields
	StreamID          uuid.UUID `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	Name              string    `sql:",notnull",validate:"required,gte=1,lte=40"` // not unique because of (project_id, lower(name)) index // note: used in stream cache
	Description       string    `validate:"omitempty,lte=255"`
	CreatedOn         time.Time `sql:",default:now()"`
	UpdatedOn         time.Time `sql:",default:now()"`
	ProjectID         uuid.UUID `sql:"on_delete:RESTRICT,notnull,type:uuid"`
	Project           *Project  `msgpack:"-"`
	AllowManualWrites bool      `sql:",notnull"`

	// Schema-related fields (note: some are used in stream cache)
	SchemaKind          StreamSchemaKind `sql:",notnull",validate:"required"`
	Schema              string           `sql:",notnull",validate:"required"`
	SchemaMD5           []byte           `sql:"schema_md5,notnull",validate:"required"`
	AvroSchema          string           `sql:",type:json,notnull",validate:"required"`
	CanonicalAvroSchema string           `sql:",type:json,notnull",validate:"required"`
	CanonicalIndexes    string           `sql:",type:json,notnull",validate:"required"`
	StreamIndexes       []*StreamIndex

	// Behaviour-related fields (note: used in stream cache)
	UseLog                    bool  `sql:",notnull"`
	UseIndex                  bool  `sql:",notnull"`
	UseWarehouse              bool  `sql:",notnull"`
	LogRetentionSeconds       int32 `sql:",notnull,default:0"`
	IndexRetentionSeconds     int32 `sql:",notnull,default:0"`
	WarehouseRetentionSeconds int32 `sql:",notnull,default:0"`

	// Instances-related fields
	StreamInstances           []*StreamInstance `msgpack:"-"`
	PrimaryStreamInstanceID   *uuid.UUID        `sql:"on_delete:SET NULL,type:uuid"`
	PrimaryStreamInstance     *StreamInstance   `msgpack:"-"`
	InstancesCreatedCount     int32             `sql:",notnull,default:0"`
	InstancesDeletedCount     int32             `sql:",notnull,default:0"`
	InstancesMadeFinalCount   int32             `sql:",notnull,default:0"`
	InstancesMadePrimaryCount int32             `sql:",notnull,default:0"`
}

// Validate runs validation on the stream
func (s *Stream) Validate() error {
	return Validator.Struct(s)
}

// GetStreamID implements engine/driver.Stream
func (s *Stream) GetStreamID() uuid.UUID {
	return s.StreamID
}

// GetStreamName implements engine/driver.Stream
func (s *Stream) GetStreamName() string {
	return s.Name
}

// GetUseLog implements engine/driver.Stream
func (s *Stream) GetUseLog() bool {
	return s.UseLog
}

// GetUseIndex implements engine/driver.Stream
func (s *Stream) GetUseIndex() bool {
	return s.UseIndex
}

// GetUseWarehouse implements engine/driver.Stream
func (s *Stream) GetUseWarehouse() bool {
	return s.UseWarehouse
}

// GetLogRetention implements engine/driver.Stream
func (s *Stream) GetLogRetention() time.Duration {
	return time.Duration(s.LogRetentionSeconds) * time.Second
}

// GetIndexRetention implements engine/driver.Stream
func (s *Stream) GetIndexRetention() time.Duration {
	return time.Duration(s.IndexRetentionSeconds) * time.Second
}

// GetWarehouseRetention implements engine/driver.Stream
func (s *Stream) GetWarehouseRetention() time.Duration {
	return time.Duration(s.WarehouseRetentionSeconds) * time.Second
}

// GetCodec implements engine/driver.Stream
func (s *Stream) GetCodec() *codec.Codec {
	if len(s.StreamIndexes) == 0 {
		panic(fmt.Errorf("GetCodec must not be called without StreamIndexes loaded"))
	}

	var primaryIndex codec.Index
	var secondaryIndexes []codec.Index
	for _, index := range s.StreamIndexes {
		if index.Primary {
			primaryIndex = index
		} else {
			secondaryIndexes = append(secondaryIndexes, index)
		}
	}

	c, err := codec.New(s.CanonicalAvroSchema, primaryIndex, secondaryIndexes)
	if err != nil {
		panic(err)
	}

	return c
}

// StreamIndex represents an index on a stream
type StreamIndex struct {
	_msgpack      struct{}  `msgpack:",omitempty"`
	tableName     struct{}  `sql:"stream_indexes"`
	StreamIndexID uuid.UUID `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	StreamID      uuid.UUID `sql:"on_delete:CASCADE,notnull,type:uuid"`
	Stream        *Stream   `msgpack:"-"`
	ShortID       int       `sql:",notnull",validate:"required"`
	Fields        []string  `sql:",notnull",validate:"required,gte=1"`
	Primary       bool      `sql:",notnull"`
	Normalize     bool      `sql:",notnull"`
}

// GetIndexID implements codec.Index
func (i *StreamIndex) GetIndexID() uuid.UUID {
	return i.StreamIndexID
}

// GetShortID implements codec.Index
func (i *StreamIndex) GetShortID() int {
	return i.ShortID
}

// GetFields implements codec.Index
func (i *StreamIndex) GetFields() []string {
	return i.Fields
}

// GetNormalize implements codec.Index
func (i *StreamIndex) GetNormalize() bool {
	return i.Normalize
}

// StreamInstance represents a single version of a stream (for a streaming stream,
// there will only be one instance; but for a batch stream, each update represents
// a new instance)
type StreamInstance struct {
	_msgpack         struct{}  `msgpack:",omitempty"`
	StreamInstanceID uuid.UUID `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	StreamID         uuid.UUID `sql:"on_delete:RESTRICT,notnull,type:uuid"`
	Stream           *Stream   `msgpack:"-"`
	CreatedOn        time.Time `sql:",default:now()"`
	UpdatedOn        time.Time `sql:",default:now()"`
	MadePrimaryOn    *time.Time
	MadeFinalOn      *time.Time
	Version          int `sql:",notnull,default:0"`
}

// EfficientStreamInstance can be used to efficiently make a UUID conform to engine/driver.StreamInstance
type EfficientStreamInstance uuid.UUID

// GetStreamInstanceID implements engine/driver.StreamInstance
func (si EfficientStreamInstance) GetStreamInstanceID() uuid.UUID {
	return uuid.UUID(si)
}

// GetStreamInstanceID implements engine/driver.StreamInstance
func (si *StreamInstance) GetStreamInstanceID() uuid.UUID {
	return si.StreamInstanceID
}

// CachedInstance contains key information about a stream for rapid (cached) lookup
type CachedInstance struct {
	StreamID                  uuid.UUID
	Public                    bool
	Final                     bool
	UseLog                    bool
	UseIndex                  bool
	UseWarehouse              bool
	LogRetentionSeconds       int32
	IndexRetentionSeconds     int32
	WarehouseRetentionSeconds int32
	OrganizationID            uuid.UUID
	OrganizationName          string
	ProjectID                 uuid.UUID
	ProjectName               string
	StreamName                string
	Codec                     *codec.Codec
}

// CachedInstance implements engine/driver.Organization, engine/driver.Project, engine/driver.StreamInstance, and engine/driver.Stream

// GetOrganizationID implements engine/driver.Organization
func (c *CachedInstance) GetOrganizationID() uuid.UUID {
	return c.OrganizationID
}

// GetOrganizationName implements engine/driver.Project
func (c *CachedInstance) GetOrganizationName() string {
	return c.OrganizationName
}

// GetProjectID implements engine/driver.Project
func (c *CachedInstance) GetProjectID() uuid.UUID {
	return c.ProjectID
}

// GetProjectName implements engine/driver.Project
func (c *CachedInstance) GetProjectName() string {
	return c.ProjectName
}

// GetPublic implements engine/driver.Project
func (c *CachedInstance) GetPublic() bool {
	return c.Public
}

// GetStreamID implements engine/driver.Stream
func (c *CachedInstance) GetStreamID() uuid.UUID {
	return c.StreamID
}

// GetStreamName implements engine/driver.Stream
func (c *CachedInstance) GetStreamName() string {
	return c.StreamName
}

// GetUseLog implements engine/driver.Stream
func (c *CachedInstance) GetUseLog() bool {
	return c.UseLog
}

// GetUseIndex implements engine/driver.Stream
func (c *CachedInstance) GetUseIndex() bool {
	return c.UseIndex
}

// GetUseWarehouse implements engine/driver.Stream
func (c *CachedInstance) GetUseWarehouse() bool {
	return c.UseWarehouse
}

// GetLogRetention implements engine/driver.Stream
func (c *CachedInstance) GetLogRetention() time.Duration {
	return time.Duration(c.LogRetentionSeconds) * time.Second
}

// GetIndexRetention implements engine/driver.Stream
func (c *CachedInstance) GetIndexRetention() time.Duration {
	return time.Duration(c.IndexRetentionSeconds) * time.Second
}

// GetWarehouseRetention implements engine/driver.Stream
func (c *CachedInstance) GetWarehouseRetention() time.Duration {
	return time.Duration(c.WarehouseRetentionSeconds) * time.Second
}

// GetCodec implements engine/driver.Stream
func (c *CachedInstance) GetCodec() *codec.Codec {
	return c.Codec
}

// ---------------
// Commands

// CreateStreamCommand contains args for creating a stream
type CreateStreamCommand struct {
	Project                   *Project
	Name                      string
	SchemaKind                StreamSchemaKind
	Schema                    string
	Indexes                   *string
	Description               *string
	AllowManualWrites         *bool
	UseLog                    *bool
	UseIndex                  *bool
	UseWarehouse              *bool
	LogRetentionSeconds       *int
	IndexRetentionSeconds     *int
	WarehouseRetentionSeconds *int
}

// UpdateStreamCommand contains args for updating a stream
type UpdateStreamCommand struct {
	Stream            *Stream
	SchemaKind        *StreamSchemaKind
	Schema            *string
	Indexes           *string
	Description       *string
	AllowManualWrites *bool
}

// ---------------
// Events

// StreamCreatedEvent is sent when a stream is created
type StreamCreatedEvent struct {
	Stream *Stream
}

// StreamUpdatedEvent is sent when a stream is updated
type StreamUpdatedEvent struct {
	Stream         *Stream
	ModifiedSchema bool
}

// StreamDeletedEvent is sent when a stream is deleted
type StreamDeletedEvent struct {
	StreamID uuid.UUID
}

// StreamInstanceCreatedEvent is sent when a stream instance is created
type StreamInstanceCreatedEvent struct {
	Stream         *Stream
	StreamInstance *StreamInstance
	MakePrimary    bool
}

// StreamInstanceUpdatedEvent is sent when a stream instance is updated (made primary or final)
type StreamInstanceUpdatedEvent struct {
	Stream         *Stream
	StreamInstance *StreamInstance
	MakeFinal      bool
	MakePrimary    bool

	// needed downstream
	OrganizationName string
	ProjectName      string
}

// StreamInstanceDeletedEvent is sent when a stream instance is deleted
type StreamInstanceDeletedEvent struct {
	Stream         *Stream
	StreamInstance *StreamInstance
}

// ---------------
// Validation

var streamNameRegex = regexp.MustCompile("^[_a-z][_a-z0-9]*$")

func init() {
	Validator.RegisterStructValidation(func(sl validator.StructLevel) {
		s := sl.Current().Interface().(Stream)

		if !streamNameRegex.MatchString(s.Name) {
			sl.ReportError(s.Name, "Name", "", "alphanumericorunderscore", "")
		}
	}, Stream{})
}
