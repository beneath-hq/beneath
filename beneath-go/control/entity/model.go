package entity

import (
	"context"
	"fmt"
	"time"

	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	uuid "github.com/satori/go.uuid"

	"github.com/beneath-core/beneath-go/db"
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
	OutputStreams []*Stream `pg:"fk:source_model_id"`
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

func init() {
	orm.RegisterTable((*StreamIntoModel)(nil))
}

// ParseModelKind returns a matching ModelKind or false if invalid
func ParseModelKind(kind string) (ModelKind, bool) {
	if kind == "streaming" {
		return ModelKindStreaming, true
	} else if kind == "batch" {
		return ModelKindBatch, true
	} else if kind == "microbatch" {
		return ModelKindMicroBatch, true
	}
	return "", false
}

// FindModel finds a model
func FindModel(ctx context.Context, modelID uuid.UUID) *Model {
	model := &Model{
		ModelID: modelID,
	}
	err := db.DB.ModelContext(ctx, model).
		Column("model.*").
		Relation("Project").
		Relation("OutputStreams").
		Relation("InputStreams").
		WherePK().
		Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return model
}

// FindModelByNameAndProject finds a model
func FindModelByNameAndProject(ctx context.Context, name string, projectName string) *Model {
	model := &Model{}
	err := db.DB.ModelContext(ctx, model).
		Column("model.*").
		Relation("Project").
		Relation("OutputStreams").
		Relation("InputStreams").
		Where("lower(model.name) = lower(?)", name).
		Where("lower(project.name) = lower(?)", projectName).
		Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return model
}

// CompileAndCreate creates the model and its dependencies
func (m *Model) CompileAndCreate(ctx context.Context, inputStreamIDs []uuid.UUID, outputStreamScheams []string) error {
	// Populate m.Project if not set
	// We do it to set Project on output streams, so they won't each look it up separately
	// They would otherwise look it up separately to get the project name to create itself in Warehouse
	if m.Project == nil {
		m.Project = FindProject(ctx, m.ProjectID)
		if m.Project == nil {
			return fmt.Errorf("Couldn't find project")
		}
	}

	// prepare output streams to create
	outputStreams := make([]*Stream, len(outputStreamScheams))
	for idx, schema := range outputStreamScheams {
		stream := &Stream{
			Schema:    schema,
			External:  false,
			Batch:     m.Kind == ModelKindBatch,
			Manual:    false,
			ProjectID: m.ProjectID,
			Project:   m.Project,
		}
		err := stream.Compile(ctx)
		if err != nil {
			return err
		}
		outputStreams[idx] = stream
	}

	// begin database transaction
	err := db.DB.WithContext(ctx).RunInTransaction(func(tx *pg.Tx) error {
		// insert model
		_, err := tx.Model(m).Insert()
		if err != nil {
			return err
		}

		// insert StreamIntoModels
		for _, streamID := range inputStreamIDs {
			_, err = tx.Model(&StreamIntoModel{
				ModelID:  m.ModelID,
				StreamID: streamID,
			}).Insert()
			if err != nil {
				return err
			}
		}

		// insert outputStreams
		for _, stream := range outputStreams {
			stream.SourceModelID = &m.ModelID
			err := stream.CreateWithTx(tx)
			if err != nil {
				return err
			}
		}

		// done
		return nil
	})

	// done
	return err
}
