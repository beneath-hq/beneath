package entity

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/go-pg/pg"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/go-playground/validator.v9"

	"github.com/beneath-core/beneath-go/core/schema"
	"github.com/beneath-core/beneath-go/db"
)

// Stream represents a collection of data
type Stream struct {
	StreamID                uuid.UUID `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	Name                    string    `sql:",notnull",validate:"required,gte=1,lte=40"` // not unique because of (project_id, user_id) index
	Description             string    `validate:"omitempty,lte=255"`
	Schema                  string    `sql:",notnull",validate:"required"`
	AvroSchema              string    `sql:",type:json,notnull",validate:"required"`
	CanonicalAvroSchema     string    `sql:",type:json,notnull",validate:"required"`
	BigQuerySchema          string    `sql:"bigquery_schema,type:json,notnull",validate:"required"`
	KeyFields               []string  `sql:",notnull",validate:"required,gte=1"`
	External                bool      `sql:",notnull"`
	Batch                   bool      `sql:",notnull"`
	Manual                  bool      `sql:",notnull"`
	ProjectID               uuid.UUID `sql:"on_delete:RESTRICT,notnull,type:uuid"`
	Project                 *Project
	ModelID                 uuid.UUID `sql:"on_delete:RESTRICT,type:uuid"`
	Model                   *Model
	DerivingModels          []*Model `pg:"many2many:streams_into_models,fk:stream_id,joinFK:model_id"`
	StreamInstances         []*StreamInstance
	CurrentStreamInstanceID *uuid.UUID `sql:"on_delete:SET NULL,type:uuid"`
	CurrentStreamInstance   *StreamInstance
	CreatedOn               time.Time `sql:",default:now()"`
	UpdatedOn               time.Time `sql:",default:now()"`
	DeletedOn               time.Time
}

var (
	// used for validation
	streamNameRegex *regexp.Regexp
)

func init() {
	// configure validation
	streamNameRegex = regexp.MustCompile("^[_a-z][_\\-a-z0-9]*$")
	GetValidator().RegisterStructValidation(streamValidation, Stream{})
}

// custom stream validation
func streamValidation(sl validator.StructLevel) {
	s := sl.Current().Interface().(Stream)

	if !streamNameRegex.MatchString(s.Name) {
		sl.ReportError(s.Name, "Name", "", "alphanumericorunderscore", "")
	}
}

// FindStream finds a stream
func FindStream(ctx context.Context, streamID uuid.UUID) *Stream {
	stream := &Stream{
		StreamID: streamID,
	}
	err := db.DB.ModelContext(ctx, stream).WherePK().Column("stream.*", "Project", "CurrentStreamInstance").Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return stream
}

// FindStreamByNameAndProject finds a stream
func FindStreamByNameAndProject(ctx context.Context, name string, projectName string) *Stream {
	stream := &Stream{}
	err := db.DB.ModelContext(ctx, stream).
		Column("stream.*", "Project", "CurrentStreamInstance").
		Where("lower(stream.name) = lower(?)", name).
		Where("lower(project.name) = lower(?)", projectName).
		Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return stream
}

// FindInstanceIDByNameAndProject returns the current instance ID of the stream
func FindInstanceIDByNameAndProject(ctx context.Context, name string, projectName string) uuid.UUID {
	return getInstanceCache().get(ctx, name, projectName)
}

// FindCachedStreamByCurrentInstanceID returns select info about the instance's stream
func FindCachedStreamByCurrentInstanceID(ctx context.Context, instanceID uuid.UUID) *CachedStream {
	return getStreamCache().get(ctx, instanceID)
}

// UpdateDetails updates a stream (only exposes fields where updates are permitted)
func (s *Stream) UpdateDetails(ctx context.Context, newSchema *string, manual *bool) error {
	if manual != nil {
		s.Manual = *manual
	}

	// we allow updating the schema with new docs/layout, but not semantic updates
	if newSchema != nil {
		// compile schema
		compiler := schema.NewCompiler(*newSchema)
		err := compiler.Compile()
		if err != nil {
			return fmt.Errorf("Error compiling schema: %s", err.Error())
		}
		streamDef := compiler.GetStream()

		// get canonical avro
		canonicalAvro, err := streamDef.BuildCanonicalAvroSchema()
		if err != nil {
			return fmt.Errorf("Error compiling schema: %s", err.Error())
		}

		// check canonical avro is the same
		if canonicalAvro != s.CanonicalAvroSchema {
			return fmt.Errorf("Unfortunately we do not currently support changing a stream's data structure; you can only edit its documentation")
		}

		// get avro schemas
		avro, err := streamDef.BuildAvroSchema()
		if err != nil {
			return fmt.Errorf("Error compiling schema: %s", err.Error())
		}

		// compute bigquery schema
		bqSchema, err := streamDef.BuildBigQuerySchema()
		if err != nil {
			return fmt.Errorf("Error compiling schema: %s", err.Error())
		}

		// set update avro and bigquery
		s.AvroSchema = avro
		s.BigQuerySchema = bqSchema
		s.Description = streamDef.Description
	}

	// validate
	err := GetValidator().Struct(s)
	if err != nil {
		return err
	}

	// TODO: Update schema in BigQuery

	// update
	_, err = db.DB.ModelContext(ctx, s).Column("description", "manual").WherePK().Update()
	return err
}

// CompileAndCreate compiles the schema, derives name and avro schemas and inserts
// the stream into the database
func (s *Stream) CompileAndCreate(ctx context.Context) error {
	// compile schema
	compiler := schema.NewCompiler(s.Schema)
	err := compiler.Compile()
	if err != nil {
		return fmt.Errorf("Error compiling schema: %s", err.Error())
	}
	streamDef := compiler.GetStream()

	// get avro schemas
	avro, err := streamDef.BuildAvroSchema()
	if err != nil {
		return fmt.Errorf("Error compiling schema: %s", err.Error())
	}

	canonicalAvro, err := streamDef.BuildCanonicalAvroSchema()
	if err != nil {
		return fmt.Errorf("Error compiling schema: %s", err.Error())
	}

	// compute bigquery schema
	bqSchema, err := streamDef.BuildBigQuerySchema()
	if err != nil {
		return fmt.Errorf("Error compiling schema: %s", err.Error())
	}

	// set missing stream fields
	s.Name = streamDef.Name
	s.KeyFields = streamDef.KeyFields
	s.AvroSchema = avro
	s.CanonicalAvroSchema = canonicalAvro
	s.BigQuerySchema = bqSchema
	s.Description = streamDef.Description

	// validate
	err = GetValidator().Struct(s)
	if err != nil {
		return err
	}

	// populate s.Project if not set
	if s.Project == nil {
		s.Project = FindProject(ctx, s.ProjectID)
	}

	// create stream (and a new stream instance ID if not batch)
	err = db.DB.WithContext(ctx).RunInTransaction(func(tx *pg.Tx) error {
		// insert stream
		_, err := tx.Model(s).Insert()
		if err != nil {
			return err
		}

		// create and set stream instance if not batch
		if !s.Batch {
			// create stream instance
			si, err := CreateStreamInstanceWithTx(tx, s.StreamID)
			if err != nil {
				return err
			}

			// update stream with stream instance ID
			s.CurrentStreamInstanceID = &si.StreamInstanceID
			_, err = tx.Model(s).WherePK().Update()
			if err != nil {
				return err
			}

			// register instance
			err = db.Engine.Warehouse.RegisterStreamInstance(
				ctx,
				s.ProjectID,
				s.Project.Name,
				s.StreamID,
				s.Name,
				s.Description,
				s.BigQuerySchema,
				s.KeyFields,
				si.StreamInstanceID,
			)
			if err != nil {
				return err
			}
		}

		// done
		return nil
	})
	if err != nil {
		return err
	}

	// done
	return nil
}
