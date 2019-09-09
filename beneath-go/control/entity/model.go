package entity

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

// Model represents a Beneath model
type Model struct {
	ModelID       uuid.UUID `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	Name          string    `sql:",notnull",validate:"required,gte=1,lte=40"` // not unique because of (project_id, model_id) index
	Description   string    `validate:"omitempty,lte=255"`
	SourceURL     string    `validate:"omitempty,url,lte=255"`
	Kind          ModelKind `sql:",notnull",validate:"required,lte=3"`
	CreatedOn     time.Time `sql:",default:now()"`
	UpdatedOn     time.Time `sql:",default:now()"`
	DeletedOn     time.Time
	ProjectID     uuid.UUID `sql:"on_delete:RESTRICT,notnull,type:uuid"`
	Project       *Project
	InputStreams  []*Stream `pg:"many2many:streams_into_models,fk:model_id,joinFK:stream_id"`
	OutputStreams []*Stream
}

// StreamIntoModel represnts the many-to-many relationship between input streams and models
type StreamIntoModel struct {
	tableName struct{}  `sql:"streams_into_models,alias:sm"`
	StreamID  uuid.UUID `sql:",pk,type:uuid"`
	Stream    *Stream
	ModelID   uuid.UUID `sql:",pk,type:uuid"`
	Model     *Model
}

// ModelKind represents a kind of model -- streaming, microbatch or batch
type ModelKind string

const (
	// ModelKindBatch is a model that replaces previous instances on update
	ModelKindBatch ModelKind = "b"

	// ModelKindMicroBatch is a model that processes streaming input periodically in chunks
	ModelKindMicroBatch ModelKind = "m"

	// ModelKindStreaming is a model that processes
	ModelKindStreaming ModelKind = "s"
)
