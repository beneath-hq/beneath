package entity

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"time"

	"github.com/go-pg/pg"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/go-playground/validator.v9"

	"github.com/beneath-core/beneath-go/core/schema"
	"github.com/beneath-core/beneath-go/db"
	"github.com/beneath-core/beneath-go/taskqueue"
	"github.com/beneath-core/beneath-go/taskqueue/task"
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
	SourceModelID           *uuid.UUID `sql:"on_delete:RESTRICT,type:uuid"`
	SourceModel             *Model
	DerivedModels           []*Model `pg:"many2many:streams_into_models,fk:stream_id,joinFK:model_id"`
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
	err := db.DB.ModelContext(ctx, stream).WherePK().Column("stream.*", "Project", "CurrentStreamInstance", "SourceModel").Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return stream
}

// FindStreamByNameAndProject finds a stream
func FindStreamByNameAndProject(ctx context.Context, name string, projectName string) *Stream {
	stream := &Stream{}
	err := db.DB.ModelContext(ctx, stream).
		Column("stream.*", "Project", "CurrentStreamInstance", "SourceModel").
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

// Compile compiles s.Schema and sets relevant fields
func (s *Stream) Compile(ctx context.Context, update bool) error {
	// compile schema
	compiler := schema.NewCompiler(s.Schema)
	err := compiler.Compile()
	if err != nil {
		return fmt.Errorf("Error compiling schema: %s", err.Error())
	}
	streamDef := compiler.GetStream()

	// get canonical avro schema
	canonicalAvro, err := streamDef.BuildCanonicalAvroSchema()
	if err != nil {
		return fmt.Errorf("Error compiling schema: %s", err.Error())
	}

	// if update, check canonical avro is the same
	if update {
		if canonicalAvro != s.CanonicalAvroSchema {
			return fmt.Errorf("Unfortunately we do not currently support changing a stream's data structure; you can only edit its documentation")
		}
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

	// check no critical changes on update
	if update {
		if s.Name != streamDef.Name {
			return fmt.Errorf("Cannot change stream name in an update")
		}
		if !reflect.DeepEqual(s.KeyFields, streamDef.KeyFields) {
			return fmt.Errorf("Cannot change stream keys in an update")
		}
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

	return nil
}

// CreateWithTx creates the stream and an associated instance (if streaming) using tx
func (s *Stream) CreateWithTx(tx *pg.Tx) error {
	// insert stream
	_, err := tx.Model(s).Insert()
	if err != nil {
		return err
	}

	// create and set stream instance if not batch
	if !s.Batch {
		// create stream instance
		si, err := s.CreateStreamInstanceWithTx(tx)
		if err != nil {
			return err
		}

		// commit instance
		err = s.CommitStreamInstanceWithTx(tx, si)
		if err != nil {
			return err
		}
	}

	// done
	return nil
}

// UpdateWithTx updates the stream (if streaming) using tx
func (s *Stream) UpdateWithTx(tx *pg.Tx) error {
	// update
	_, err := tx.Model(s).WherePK().Update()
	if err != nil {
		return err
	}

	// update in bigquery
	if s.CurrentStreamInstanceID != nil {
		err = db.Engine.Warehouse.UpdateStreamInstance(
			tx.Context(),
			s.Project.Name,
			s.Name,
			s.Description,
			s.BigQuerySchema,
			*s.CurrentStreamInstanceID,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// CompileAndCreate compiles the schema, derives name and avro schemas and inserts
// the stream into the database
func (s *Stream) CompileAndCreate(ctx context.Context) error {
	// compile
	err := s.Compile(ctx, false)
	if err != nil {
		return err
	}

	// create stream (and a new stream instance ID if not batch)
	err = db.DB.WithContext(ctx).RunInTransaction(func(tx *pg.Tx) error {
		return s.CreateWithTx(tx)
	})
	if err != nil {
		return err
	}

	return nil
}

// CompileAndUpdate updates a stream.
// We allow updating the schema with new docs/layout, but not semantic updates.
func (s *Stream) CompileAndUpdate(ctx context.Context, newSchema *string, manual *bool) error {
	if manual != nil {
		s.Manual = *manual
	}

	if newSchema != nil {
		s.Schema = *newSchema
		s.Compile(ctx, true)
	}

	return db.DB.WithContext(ctx).RunInTransaction(func(tx *pg.Tx) error {
		return s.UpdateWithTx(tx)
	})
}

// Delete deletes a stream and all its related instances
func (s *Stream) Delete(ctx context.Context) error {
	// get instances
	var instances []*StreamInstance
	err := db.DB.ModelContext(ctx, &instances).
		Where("stream_id = ?", s.StreamID).
		Select()
	if err != nil {
		return err
	}

	// delete instances
	for _, inst := range instances {
		err := s.DeleteStreamInstance(ctx, inst)
		if err != nil {
			return err
		}
	}

	// delete stream
	err = db.DB.WithContext(ctx).Delete(s)
	if err != nil {
		return err
	}

	return nil
}

// CreateStreamInstance creates a new instance
func (s *Stream) CreateStreamInstance(ctx context.Context) (res *StreamInstance, err error) {
	err = db.DB.WithContext(ctx).RunInTransaction(func(tx *pg.Tx) error {
		res, err = s.CreateStreamInstanceWithTx(tx)
		return err
	})
	return res, err
}

// CreateStreamInstanceWithTx is the same as CreateStreamInstance, but in a database transaction
func (s *Stream) CreateStreamInstanceWithTx(tx *pg.Tx) (*StreamInstance, error) {
	// check uncommited instances count
	var count int
	_, err := tx.QueryOne(pg.Scan(&count), `
		select count(*)
		from streams s
		join stream_instances si on s.stream_id = si.stream_id
		where s.stream_id = ?
		and s.current_stream_instance_id is distinct from si.stream_instance_id`, s.StreamID)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, fmt.Errorf("Another batch is already outstanding for stream '%s/%s' – commit or clear it before continuing", s.Project.Name, s.Name)
	}

	// create new
	si := &StreamInstance{StreamID: s.StreamID}
	_, err = tx.Model(si).Insert()
	if err != nil {
		return nil, err
	}
	si.Stream = s

	// register instance
	err = db.Engine.Warehouse.RegisterStreamInstance(
		tx.Context(),
		s.Project.Name,
		s.StreamID,
		s.Name,
		s.Description,
		s.BigQuerySchema,
		s.KeyFields,
		si.StreamInstanceID,
	)
	if err != nil {
		return nil, err
	}

	return si, nil
}

// CommitStreamInstance promotes an instance to current_instance_id and deletes the old instance
func (s *Stream) CommitStreamInstance(ctx context.Context, instance *StreamInstance) error {
	return db.DB.WithContext(ctx).RunInTransaction(func(tx *pg.Tx) error {
		return s.CommitStreamInstanceWithTx(tx, instance)
	})
}

// CommitStreamInstanceWithTx is the same as CommitStreamInstance, but in a database transaction
// Note must support multiple calls (idempotence)
func (s *Stream) CommitStreamInstanceWithTx(tx *pg.Tx, instance *StreamInstance) error {
	// check
	if instance.StreamID != s.StreamID {
		return fmt.Errorf("Cannot commit instance '%s' because it doesn't belong to stream '%s'", instance.StreamInstanceID.String(), s.StreamID.String())
	}

	// update stream with stream instance ID
	prevInstanceID := s.CurrentStreamInstanceID
	s.CurrentStreamInstanceID = &instance.StreamInstanceID
	_, err := tx.Model(s).WherePK().Update()
	if err != nil {
		return err
	}

	// call on warehouse
	err = db.Engine.Warehouse.PromoteStreamInstance(
		tx.Context(),
		s.Project.Name,
		s.StreamID,
		s.Name,
		s.Description,
		instance.StreamInstanceID,
	)
	if err != nil {
		return err
	}

	// delete old instance (unless idempotence)
	if prevInstanceID != nil && *prevInstanceID != instance.StreamInstanceID {
		err := s.DeleteStreamInstanceWithTx(tx, &StreamInstance{
			StreamID:         s.StreamID,
			StreamInstanceID: *prevInstanceID,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteStreamInstance deletes and deregisters a stream instance
func (s *Stream) DeleteStreamInstance(ctx context.Context, si *StreamInstance) error {
	return db.DB.WithContext(ctx).RunInTransaction(func(tx *pg.Tx) error {
		return s.DeleteStreamInstanceWithTx(tx, si)
	})
}

// DeleteStreamInstanceWithTx is like DeleteStreamInstance in a tx
func (s *Stream) DeleteStreamInstanceWithTx(tx *pg.Tx, si *StreamInstance) error {
	// check
	if si.StreamID != s.StreamID {
		return fmt.Errorf("Cannot delete instance '%s' because it doesn't belong to stream '%s'", si.StreamInstanceID.String(), s.StreamID.String())
	}

	// remove as current stream instance (if necessary)
	if s.CurrentStreamInstanceID != nil && *s.CurrentStreamInstanceID == si.StreamInstanceID {
		s.CurrentStreamInstanceID = nil
		_, err := tx.Model(s).Column("current_stream_instance_id").WherePK().Update()
		if err != nil {
			return err
		}
	}

	// delete
	err := tx.Delete(si)
	if err != nil {
		return err
	}

	// deregister
	err = taskqueue.Submit(tx.Context(), &task.CleanupInstance{
		InstanceID:  si.StreamInstanceID,
		StreamID:    s.StreamID,
		StreamName:  s.Name,
		ProjectID:   s.ProjectID,
		ProjectName: s.Project.Name,
	})
	if err != nil {
		return err
	}

	return nil
}

// ClearPendingBatches clears instance IDs that are not current
func (s *Stream) ClearPendingBatches(ctx context.Context) error {
	// get instances
	if len(s.StreamInstances) == 0 {
		err := db.DB.ModelContext(ctx, (*StreamInstance)(nil)).
			Where("stream_id = ?", s.StreamID).
			Select(&s.StreamInstances)
		if err != nil {
			return err
		}
	}

	// loop
	for _, si := range s.StreamInstances {
		if s.CurrentStreamInstanceID == nil || *s.CurrentStreamInstanceID != si.StreamInstanceID {
			err := s.DeleteStreamInstance(ctx, si)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
