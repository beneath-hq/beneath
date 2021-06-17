package models

import (
	"fmt"
	"regexp"
	"time"

	uuid "github.com/satori/go.uuid"
	"gopkg.in/go-playground/validator.v9"

	"github.com/beneath-hq/beneath/pkg/codec"
)

// TableSchemaKind indicates the SDL of a table's schema
type TableSchemaKind string

// Constants for TableSchemaKind
const (
	TableSchemaKindGraphQL TableSchemaKind = "GraphQL"
	TableSchemaKindAvro    TableSchemaKind = "Avro"
)

// Table represents a collection of data
type Table struct {
	_msgpack struct{} `msgpack:",omitempty"`

	// Descriptive fields
	TableID           uuid.UUID `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	Name              string    `sql:",notnull",validate:"required,gte=1,lte=40"` // not unique because of (project_id, lower(name)) index // note: used in table cache
	Description       string    `validate:"omitempty,lte=255"`
	CreatedOn         time.Time `sql:",default:now()"`
	UpdatedOn         time.Time `sql:",default:now()"`
	ProjectID         uuid.UUID `sql:"on_delete:RESTRICT,notnull,type:uuid"`
	Project           *Project  `msgpack:"-"`
	Meta              bool      `sql:",notnull"`
	AllowManualWrites bool      `sql:",notnull"`

	// Schema-related fields (note: some are used in table cache)
	SchemaKind          TableSchemaKind `sql:",notnull",validate:"required"`
	Schema              string          `sql:",notnull",validate:"required"`
	SchemaMD5           []byte          `sql:"schema_md5,notnull",validate:"required"`
	AvroSchema          string          `sql:",type:json,notnull",validate:"required"`
	CanonicalAvroSchema string          `sql:",type:json,notnull",validate:"required"`
	CanonicalIndexes    string          `sql:",type:json,notnull",validate:"required"`
	TableIndexes        []*TableIndex

	// Behaviour-related fields (note: used in table cache)
	UseLog                    bool  `sql:",notnull"`
	UseIndex                  bool  `sql:",notnull"`
	UseWarehouse              bool  `sql:",notnull"`
	LogRetentionSeconds       int32 `sql:",notnull,default:0"`
	IndexRetentionSeconds     int32 `sql:",notnull,default:0"`
	WarehouseRetentionSeconds int32 `sql:",notnull,default:0"`

	// Instances-related fields
	TableInstances            []*TableInstance `msgpack:"-"`
	PrimaryTableInstanceID    *uuid.UUID       `sql:"on_delete:SET NULL,type:uuid"`
	PrimaryTableInstance      *TableInstance   `msgpack:"-"`
	NextInstanceVersion       int              `sql:",notnull,default:0"`
	InstancesCreatedCount     int32            `sql:",notnull,default:0"`
	InstancesDeletedCount     int32            `sql:",notnull,default:0"`
	InstancesMadeFinalCount   int32            `sql:",notnull,default:0"`
	InstancesMadePrimaryCount int32            `sql:",notnull,default:0"`
}

// Validate runs validation on the table
func (s *Table) Validate() error {
	return Validator.Struct(s)
}

// GetTableID implements engine/driver.Table
func (s *Table) GetTableID() uuid.UUID {
	return s.TableID
}

// GetTableName implements engine/driver.Table
func (s *Table) GetTableName() string {
	return s.Name
}

// GetUseLog implements engine/driver.Table
func (s *Table) GetUseLog() bool {
	return s.UseLog
}

// GetUseIndex implements engine/driver.Table
func (s *Table) GetUseIndex() bool {
	return s.UseIndex
}

// GetUseWarehouse implements engine/driver.Table
func (s *Table) GetUseWarehouse() bool {
	return s.UseWarehouse
}

// GetLogRetention implements engine/driver.Table
func (s *Table) GetLogRetention() time.Duration {
	return time.Duration(s.LogRetentionSeconds) * time.Second
}

// GetIndexRetention implements engine/driver.Table
func (s *Table) GetIndexRetention() time.Duration {
	return time.Duration(s.IndexRetentionSeconds) * time.Second
}

// GetWarehouseRetention implements engine/driver.Table
func (s *Table) GetWarehouseRetention() time.Duration {
	return time.Duration(s.WarehouseRetentionSeconds) * time.Second
}

// GetCodec implements engine/driver.Table
func (s *Table) GetCodec() *codec.Codec {
	if len(s.TableIndexes) == 0 {
		panic(fmt.Errorf("GetCodec must not be called without TableIndexes loaded"))
	}

	var primaryIndex codec.Index
	var secondaryIndexes []codec.Index
	for _, index := range s.TableIndexes {
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

// TableIndex represents an index on a table
type TableIndex struct {
	_msgpack     struct{}  `msgpack:",omitempty"`
	tableName    struct{}  `sql:"table_indexes"`
	TableIndexID uuid.UUID `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	TableID      uuid.UUID `sql:"on_delete:CASCADE,notnull,type:uuid"`
	Table        *Table    `msgpack:"-"`
	ShortID      int       `sql:",notnull",validate:"required"`
	Fields       []string  `sql:",notnull",validate:"required,gte=1"`
	Primary      bool      `sql:",notnull"`
	Normalize    bool      `sql:",notnull"`
}

// GetIndexID implements codec.Index
func (i *TableIndex) GetIndexID() uuid.UUID {
	return i.TableIndexID
}

// GetShortID implements codec.Index
func (i *TableIndex) GetShortID() int {
	return i.ShortID
}

// GetFields implements codec.Index
func (i *TableIndex) GetFields() []string {
	return i.Fields
}

// GetNormalize implements codec.Index
func (i *TableIndex) GetNormalize() bool {
	return i.Normalize
}

// TableInstance represents a single version of a table (for a tableing table,
// there will only be one instance; but for a batch table, each update represents
// a new instance)
type TableInstance struct {
	_msgpack        struct{}  `msgpack:",omitempty"`
	TableInstanceID uuid.UUID `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	TableID         uuid.UUID `sql:"on_delete:RESTRICT,notnull,type:uuid"`
	Table           *Table    `msgpack:"-"`
	CreatedOn       time.Time `sql:",default:now()"`
	UpdatedOn       time.Time `sql:",default:now()"`
	MadePrimaryOn   *time.Time
	MadeFinalOn     *time.Time
	Version         int `sql:",notnull,default:0"`
}

// EfficientTableInstance can be used to efficiently make a UUID conform to engine/driver.TableInstance
type EfficientTableInstance uuid.UUID

// GetTableInstanceID implements engine/driver.TableInstance
func (si EfficientTableInstance) GetTableInstanceID() uuid.UUID {
	return uuid.UUID(si)
}

// GetTableInstanceID implements engine/driver.TableInstance
func (si *TableInstance) GetTableInstanceID() uuid.UUID {
	return si.TableInstanceID
}

// CachedInstance contains key information about a table for rapid (cached) lookup
type CachedInstance struct {
	TableID                   uuid.UUID
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
	TableName                 string
	Codec                     *codec.Codec
}

// CachedInstance implements engine/driver.Organization, engine/driver.Project, engine/driver.TableInstance, and engine/driver.Table

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

// GetTableID implements engine/driver.Table
func (c *CachedInstance) GetTableID() uuid.UUID {
	return c.TableID
}

// GetTableName implements engine/driver.Table
func (c *CachedInstance) GetTableName() string {
	return c.TableName
}

// GetUseLog implements engine/driver.Table
func (c *CachedInstance) GetUseLog() bool {
	return c.UseLog
}

// GetUseIndex implements engine/driver.Table
func (c *CachedInstance) GetUseIndex() bool {
	return c.UseIndex
}

// GetUseWarehouse implements engine/driver.Table
func (c *CachedInstance) GetUseWarehouse() bool {
	return c.UseWarehouse
}

// GetLogRetention implements engine/driver.Table
func (c *CachedInstance) GetLogRetention() time.Duration {
	return time.Duration(c.LogRetentionSeconds) * time.Second
}

// GetIndexRetention implements engine/driver.Table
func (c *CachedInstance) GetIndexRetention() time.Duration {
	return time.Duration(c.IndexRetentionSeconds) * time.Second
}

// GetWarehouseRetention implements engine/driver.Table
func (c *CachedInstance) GetWarehouseRetention() time.Duration {
	return time.Duration(c.WarehouseRetentionSeconds) * time.Second
}

// GetCodec implements engine/driver.Table
func (c *CachedInstance) GetCodec() *codec.Codec {
	return c.Codec
}

// ---------------
// Commands

// CreateTableCommand contains args for creating a table
type CreateTableCommand struct {
	Project                   *Project
	Name                      string
	SchemaKind                TableSchemaKind
	Schema                    string
	Indexes                   *string
	Description               *string
	Meta                      *bool
	AllowManualWrites         *bool
	UseLog                    *bool
	UseIndex                  *bool
	UseWarehouse              *bool
	LogRetentionSeconds       *int
	IndexRetentionSeconds     *int
	WarehouseRetentionSeconds *int
}

// UpdateTableCommand contains args for updating a table
type UpdateTableCommand struct {
	Table             *Table
	SchemaKind        *TableSchemaKind
	Schema            *string
	Indexes           *string
	Description       *string
	Meta              *bool
	AllowManualWrites *bool
}

// ---------------
// Events

// TableCreatedEvent is sent when a table is created
type TableCreatedEvent struct {
	Table *Table
}

// TableUpdatedEvent is sent when a table is updated
type TableUpdatedEvent struct {
	Table          *Table
	ModifiedSchema bool
}

// TableDeletedEvent is sent when a table is deleted
type TableDeletedEvent struct {
	TableID uuid.UUID
}

// TableInstanceCreatedEvent is sent when a table instance is created
type TableInstanceCreatedEvent struct {
	Table         *Table
	TableInstance *TableInstance
	MakePrimary   bool
}

// TableInstanceUpdatedEvent is sent when a table instance is updated (made primary or final)
type TableInstanceUpdatedEvent struct {
	Table         *Table
	TableInstance *TableInstance
	MakeFinal     bool
	MakePrimary   bool

	// needed downstream
	OrganizationName string
	ProjectName      string
}

// TableInstanceDeletedEvent is sent when a table instance is deleted
type TableInstanceDeletedEvent struct {
	Table         *Table
	TableInstance *TableInstance
}

// ---------------
// Validation

var tableNameRegex = regexp.MustCompile("^[_a-z][_a-z0-9]*$")

func init() {
	Validator.RegisterStructValidation(func(sl validator.StructLevel) {
		s := sl.Current().Interface().(Table)

		if !tableNameRegex.MatchString(s.Name) {
			sl.ReportError(s.Name, "Name", "", "alphanumericorunderscore", "")
		}
	}, Table{})
}
