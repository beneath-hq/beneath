package entity

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"reflect"
	"regexp"
	"time"

	"github.com/go-pg/pg/v9"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/go-playground/validator.v9"

	"gitlab.com/beneath-hq/beneath/control/taskqueue"
	"gitlab.com/beneath-hq/beneath/internal/hub"
	"gitlab.com/beneath-hq/beneath/pkg/codec"
	"gitlab.com/beneath-hq/beneath/pkg/schema"
)

// Stream represents a collection of data
type Stream struct {
	// Descriptive fields
	StreamID      uuid.UUID `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	Name          string    `sql:",notnull",validate:"required,gte=1,lte=40"` // not unique because of (project_id, lower(name)) index // note: used in stream cache
	Description   string    `validate:"omitempty,lte=255"`
	CreatedOn     time.Time `sql:",default:now()"`
	UpdatedOn     time.Time `sql:",default:now()"`
	ProjectID     uuid.UUID `sql:"on_delete:RESTRICT,notnull,type:uuid"`
	Project       *Project
	SourceModelID *uuid.UUID `sql:"on_delete:RESTRICT,type:uuid"`
	SourceModel   *Model
	DerivedModels []*Model `pg:"many2many:streams_into_models,fk:stream_id,joinFK:model_id"`

	// Schema-related fields (note: some are used in stream cache)
	SchemaKind          StreamSchemaKind `sql:",notnull",validate:"required"`
	Schema              string           `sql:",notnull",validate:"required"`
	SchemaMD5           []byte           `sql:"schema_md5,notnull",validate:"required"`
	AvroSchema          string           `sql:",type:json,notnull",validate:"required"`
	CanonicalAvroSchema string           `sql:",type:json,notnull",validate:"required"`
	BigQuerySchema      string           `sql:"bigquery_schema,type:json,notnull",validate:"required"`
	StreamIndexes       []*StreamIndex

	// Behaviour-related fields (note: used in stream cache)
	RetentionSeconds   int32 `sql:",notnull,default:0"`
	EnableManualWrites bool  `sql:",notnull"`

	// Instances-related fields
	StreamInstances           []*StreamInstance
	PrimaryStreamInstanceID   *uuid.UUID `sql:"on_delete:SET NULL,type:uuid"`
	PrimaryStreamInstance     *StreamInstance
	InstancesCreatedCount     int32 `sql:",notnull,default:0"`
	InstancesDeletedCount     int32 `sql:",notnull,default:0"`
	InstancesMadeFinalCount   int32 `sql:",notnull,default:0"`
	InstancesMadePrimaryCount int32 `sql:",notnull,default:0"`
}

// StreamSchemaKind indicates the SDL of a stream's schema
type StreamSchemaKind string

const (
	// StreamSchemaKindGraphQl is the GraphQL SDL
	StreamSchemaKindGraphQl StreamSchemaKind = "GraphQL"
)

var (
	// used for validation
	streamNameRegex *regexp.Regexp
)

func init() {
	// configure validation
	streamNameRegex = regexp.MustCompile("^[_a-z][_a-z0-9]*$")
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
	err := hub.DB.ModelContext(ctx, stream).
		WherePK().
		Column(
			"stream.*",
			"Project",
			"Project.Organization",
			"PrimaryStreamInstance",
			"SourceModel",
			"StreamIndexes",
		).
		Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return stream
}

// FindStreamByOrganizationProjectAndName finds a stream
func FindStreamByOrganizationProjectAndName(ctx context.Context, organizationName string, projectName string, streamName string) *Stream {
	stream := &Stream{}
	err := hub.DB.ModelContext(ctx, stream).
		Column(
			"stream.*",
			"Project",
			"Project.Organization",
			"PrimaryStreamInstance",
			"SourceModel",
			"StreamIndexes",
		).
		Where("lower(project__organization.name) = lower(?)", organizationName).
		Where("lower(project.name) = lower(?)", projectName).
		Where("lower(stream.name) = lower(?)", streamName).
		Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return stream
}

// GetStreamID implements engine/driver.Stream
func (s *Stream) GetStreamID() uuid.UUID {
	return s.StreamID
}

// GetStreamName implements engine/driver.Stream
func (s *Stream) GetStreamName() string {
	return s.Name
}

// GetRetention implements engine/driver.Stream
func (s *Stream) GetRetention() time.Duration {
	return time.Duration(s.RetentionSeconds) * time.Second
}

// GetCodec implements engine/driver.Stream
func (s *Stream) GetCodec() *codec.Codec {
	if len(s.StreamIndexes) == 0 {
		panic(fmt.Errorf("GetCodec must not be called withut StreamIndexes loaded"))
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

// GetBigQuerySchema implements engine/driver.Stream
func (s *Stream) GetBigQuerySchema() string {
	return s.BigQuerySchema
}

// Compile compiles the given schema and sets relevant fields
func (s *Stream) Compile(ctx context.Context, schemaKind StreamSchemaKind, newSchema string, update bool) error {
	// checck schema kind is gql
	if schemaKind != StreamSchemaKindGraphQl {
		return fmt.Errorf("Unsupported stream schema definition language '%s'", schemaKind)
	}

	// compile schema
	compiler := schema.NewCompiler(newSchema)
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
		if !s.indexesEqual(streamDef) {
			return fmt.Errorf("Cannot change stream key or indexes in an update")
		}
	}

	// TEMPORARY: disable secondary indexes
	if len(streamDef.SecondaryIndexes) != 0 {
		return fmt.Errorf("Cannot add secondary indexes to stream (a stream must have exactly one @key index)")
	}

	// set indexes
	if !update {
		s.assignIndexes(streamDef)
	}

	// set missing stream fields
	s.SchemaKind = schemaKind
	s.Schema = newSchema
	s.AvroSchema = avro
	s.CanonicalAvroSchema = canonicalAvro
	s.BigQuerySchema = bqSchema
	s.Description = streamDef.Description

	// validate
	err = GetValidator().Struct(s)
	if err != nil {
		return err
	}

	return nil
}

// Sets StreamIndexes to new StreamIndex objects based on def.
// Doesn't execute any DB actions, so doesn't set any IDs.
func (s *Stream) assignIndexes(def *schema.StreamDef) {
	primary := &StreamIndex{
		ShortID:   0,
		Fields:    def.KeyIndex.Fields,
		Primary:   true,
		Normalize: def.KeyIndex.Normalize,
	}

	s.StreamIndexes = []*StreamIndex{primary}

	for idx, index := range def.SecondaryIndexes {
		s.StreamIndexes = append(s.StreamIndexes, &StreamIndex{
			ShortID:   idx + 1,
			Fields:    index.Fields,
			Primary:   false,
			Normalize: index.Normalize,
		})
	}
}

// Returns true if the indexes defined in def match
func (s *Stream) indexesEqual(def *schema.StreamDef) bool {
	if len(s.StreamIndexes) != len(def.SecondaryIndexes)+1 {
		return false
	}

	equal := true
	for _, index := range s.StreamIndexes {
		if index.Primary {
			equal = equal && index.Normalize == def.KeyIndex.Normalize
			equal = equal && reflect.DeepEqual(index.Fields, def.KeyIndex.Fields)
		} else {
			found := false
			for _, secondary := range def.SecondaryIndexes {
				if index.Normalize == secondary.Normalize {
					if reflect.DeepEqual(index.Fields, secondary.Fields) {
						found = true
						break
					}
				}
			}
			equal = equal && found
		}
		if !equal {
			break
		}
	}

	return equal
}

// Stage creates or updates the stream so that it adheres to the params (or does nothing if everything already conforms).
// If s is a new stream object, Name and ProjectID must be set before calling Stage.
func (s *Stream) Stage(ctx context.Context, schemaKind StreamSchemaKind, schema string, retentionSeconds *int, enableManualWrites *bool, createPrimaryStreamInstance *bool) error {
	// determine whether to insert or update
	update := (s.StreamID != uuid.Nil)

	// tracks whether a save is necessary
	save := !update

	// compile if necessary
	schemaMD5 := md5.Sum([]byte(schema))
	if schemaKind != s.SchemaKind || !bytes.Equal(schemaMD5[:], s.SchemaMD5) {
		err := s.Compile(ctx, schemaKind, schema, update)
		if err != nil {
			return err
		}
		s.SchemaMD5 = schemaMD5[:]
		save = true
	}

	// update fields if necessary
	if retentionSeconds != nil && int32(*retentionSeconds) != s.RetentionSeconds {
		s.RetentionSeconds = int32(*retentionSeconds)
		save = true
	}
	if enableManualWrites != nil && *enableManualWrites != s.EnableManualWrites {
		s.EnableManualWrites = *enableManualWrites
		save = true
	}

	// quit if no changes
	if !save {
		return nil
	}

	// populate s.Project and s.Project.Organization if not set (i.e. due to inserting new stream)
	if s.Project == nil {
		s.Project = FindProject(ctx, s.ProjectID)
	}

	// insert or update
	return hub.DB.WithContext(ctx).RunInTransaction(func(tx *pg.Tx) error {
		if update {
			// update
			s.UpdatedOn = time.Now()
			_, err := tx.Model(s).WherePK().Update()
			if err != nil {
				return err
			}

			// update instances in engine and clear cached instances
			instances := FindStreamInstances(tx.Context(), s.StreamID)
			for _, instance := range instances {
				getStreamCache().Clear(tx.Context(), instance.StreamInstanceID)
				err := hub.Engine.RegisterInstance(tx.Context(), s.Project, s, instance)
				if err != nil {
					return err
				}
			}

			// done with updates
			return nil
		}

		// IT'S AN INSERT (NOT UPDATE)

		// insert
		_, err := tx.Model(s).Insert()
		if err != nil {
			return err
		}

		// create indexes
		for _, index := range s.StreamIndexes {
			index.StreamID = s.StreamID
			_, err := tx.Model(index).Insert()
			if err != nil {
				return err
			}
		}

		// create a primary instance if requested
		if createPrimaryStreamInstance != nil && *createPrimaryStreamInstance {
			// create stream instance
			instance, err := s.CreateStreamInstanceWithTx(tx)
			if err != nil {
				return err
			}

			// commit instance
			err = s.UpdateStreamInstanceWithTx(tx, instance, false, true, false)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// Delete deletes a stream and all its related instances
func (s *Stream) Delete(ctx context.Context) error {
	// delete instances
	instances := FindStreamInstances(ctx, s.StreamID)
	for _, inst := range instances {
		err := s.DeleteStreamInstance(ctx, inst)
		if err != nil {
			return err
		}
	}

	// delete stream
	err := hub.DB.WithContext(ctx).Delete(s)
	if err != nil {
		return err
	}

	// background cleanup
	err = taskqueue.Submit(ctx, &CleanupStreamTask{
		StreamID: s.StreamID,
	})
	if err != nil {
		return err
	}

	return nil
}

// CreateStreamInstance creates a new instance
func (s *Stream) CreateStreamInstance(ctx context.Context) (res *StreamInstance, err error) {
	err = hub.DB.WithContext(ctx).RunInTransaction(func(tx *pg.Tx) error {
		res, err = s.CreateStreamInstanceWithTx(tx)
		return err
	})
	return res, err
}

// CreateStreamInstanceWithTx is the same as CreateStreamInstance, but in a database transaction
func (s *Stream) CreateStreamInstanceWithTx(tx *pg.Tx) (*StreamInstance, error) {
	// create new
	si := &StreamInstance{
		StreamID: s.StreamID,
		Stream:   s,
	}
	_, err := tx.Model(si).Insert()
	if err != nil {
		return nil, err
	}

	// increment count
	s.InstancesCreatedCount++
	_, err = tx.Model(s).Set("instances_created_count = instances_created_count + 1").WherePK().Update()
	if err != nil {
		return nil, err
	}

	// register instance
	err = hub.Engine.RegisterInstance(tx.Context(), s.Project, s, si)
	if err != nil {
		return nil, err
	}

	return si, nil
}

// UpdateStreamInstance updates the state of a stream instance
func (s *Stream) UpdateStreamInstance(ctx context.Context, instance *StreamInstance, makeFinal bool, makePrimary bool, deletePreviousPrimary bool) error {
	return hub.DB.WithContext(ctx).RunInTransaction(func(tx *pg.Tx) error {
		return s.UpdateStreamInstanceWithTx(tx, instance, makeFinal, makePrimary, deletePreviousPrimary)
	})
}

// UpdateStreamInstanceWithTx is like UpdateStreamInstance, but runs in a transaction
func (s *Stream) UpdateStreamInstanceWithTx(tx *pg.Tx, instance *StreamInstance, makeFinal bool, makePrimary bool, deletePreviousPrimary bool) error {
	// check relationship
	if instance.StreamID != s.StreamID {
		return fmt.Errorf("Cannot update instance '%s' because it doesn't belong to stream '%s'", instance.StreamInstanceID.String(), s.StreamID.String())
	}

	// check makePrimary and deletePreviousPrimary connection
	if !makePrimary && deletePreviousPrimary {
		return fmt.Errorf("Cannot set deletePreviousPrimary if makePrimary isn't set")
	}

	// bail if there's nothing to do
	if (!makeFinal || instance.MadeFinalOn != nil) && !makePrimary {
		return nil
	}

	// we know there's something to do

	// prepare models to update
	sm := tx.Model(s)
	im := tx.Model(instance)

	// values needed later
	now := time.Now()
	previousPrimaryInstanceID := s.PrimaryStreamInstanceID

	// handle makeFinal
	if makeFinal && instance.MadeFinalOn == nil {
		s.InstancesMadeFinalCount++
		sm = sm.Set("instances_made_final_count = instances_made_final_count + 1")
		instance.MadeFinalOn = &now
		im = im.Set("made_final_on = now()")
	}

	// handle makePrimary
	if makePrimary {
		s.PrimaryStreamInstanceID = &instance.StreamInstanceID
		sm = sm.Set("primary_stream_instance_id = ?", instance.StreamInstanceID)
		if instance.MadePrimaryOn == nil {
			s.InstancesMadePrimaryCount++
			sm = sm.Set("instances_made_primary_count = instances_made_primary_count + 1")
			instance.MadePrimaryOn = &now
			im = im.Set("made_primary_on = now()")
		}
	}

	// updated on timestamps
	s.UpdatedOn = now
	sm = sm.Set("updated_on = now()")
	instance.UpdatedOn = now
	im = im.Set("updated_on = now()")

	// update
	_, err := sm.WherePK().Update()
	if err != nil {
		return err
	}
	_, err = im.WherePK().Update()
	if err != nil {
		return err
	}

	// update in warehouse
	if makePrimary {
		err = hub.Engine.PromoteInstance(tx.Context(), s.Project, s, instance)
		if err != nil {
			return err
		}
	}

	// delete previous primary instance
	if deletePreviousPrimary && previousPrimaryInstanceID != nil {
		err := s.DeleteStreamInstanceWithTx(tx, &StreamInstance{
			StreamInstanceID: *previousPrimaryInstanceID,
			StreamID:         s.StreamID,
		})
		if err != nil {
			return err
		}
	}

	// clear caches
	getStreamCache().Clear(tx.Context(), instance.StreamInstanceID)
	if makePrimary {
		getInstanceCache().clear(tx.Context(), s.Project.Organization.Name, s.Project.Name, s.Name)
		if previousPrimaryInstanceID != nil {
			getStreamCache().Clear(tx.Context(), *previousPrimaryInstanceID)
		}
	}

	return nil
}

// DeleteStreamInstance deletes and deregisters a stream instance
func (s *Stream) DeleteStreamInstance(ctx context.Context, si *StreamInstance) error {
	return hub.DB.WithContext(ctx).RunInTransaction(func(tx *pg.Tx) error {
		return s.DeleteStreamInstanceWithTx(tx, si)
	})
}

// DeleteStreamInstanceWithTx is like DeleteStreamInstance in a tx
func (s *Stream) DeleteStreamInstanceWithTx(tx *pg.Tx, instance *StreamInstance) error {
	// check relationship
	if instance.StreamID != s.StreamID {
		return fmt.Errorf("Cannot delete instance '%s' because it doesn't belong to stream '%s'", instance.StreamInstanceID.String(), s.StreamID.String())
	}

	// delete instance
	err := tx.Delete(instance)
	if err != nil {
		return err
	}

	// clear cache
	getStreamCache().Clear(tx.Context(), instance.StreamInstanceID)

	// if we deleted the primary instance, clear it (not part of stream update because of SET NULL constraint)
	if s.PrimaryStreamInstanceID != nil && *s.PrimaryStreamInstanceID == instance.StreamInstanceID {
		s.PrimaryStreamInstanceID = nil
		s.PrimaryStreamInstance = nil
		getInstanceCache().clear(tx.Context(), s.Project.Organization.Name, s.Project.Name, s.Name)
	}

	// update stream
	s.UpdatedOn = time.Now()
	s.InstancesDeletedCount++
	_, err = tx.Model(s).
		Set("updated_on = ?", s.UpdatedOn).
		Set("instances_deleted_count = instances_deleted_count + 1").
		WherePK().
		Update()
	if err != nil {
		return err
	}

	// deregister instance
	err = taskqueue.Submit(tx.Context(), &CleanupInstanceTask{
		InstanceID:   instance.StreamInstanceID,
		CachedStream: NewCachedStream(s, instance),
	})
	if err != nil {
		return err
	}

	return nil
}
