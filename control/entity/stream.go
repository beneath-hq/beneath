package entity

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/go-pg/pg/v9"
	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/gqlerror"
	"gopkg.in/go-playground/validator.v9"

	"gitlab.com/beneath-hq/beneath/control/taskqueue"
	"gitlab.com/beneath-hq/beneath/internal/hub"
	"gitlab.com/beneath-hq/beneath/pkg/codec"
	"gitlab.com/beneath-hq/beneath/pkg/schemalang"
	"gitlab.com/beneath-hq/beneath/pkg/schemalang/transpilers"
)

// Stream represents a collection of data
type Stream struct {
	// Descriptive fields
	StreamID          uuid.UUID `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	Name              string    `sql:",notnull",validate:"required,gte=1,lte=40"` // not unique because of (project_id, lower(name)) index // note: used in stream cache
	Description       string    `validate:"omitempty,lte=255"`
	CreatedOn         time.Time `sql:",default:now()"`
	UpdatedOn         time.Time `sql:",default:now()"`
	ProjectID         uuid.UUID `sql:"on_delete:RESTRICT,notnull,type:uuid"`
	Project           *Project
	AllowManualWrites bool `sql:",notnull"`

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

// Constants for StreamSchemaKind
const (
	StreamSchemaKindGraphQL StreamSchemaKind = "GraphQL"
	StreamSchemaKindAvro    StreamSchemaKind = "Avro"
)

// MaxInstancesPerStream sets a limit for the number of instances for a stream at any given time
const MaxInstancesPerStream = 25

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

// Compile compiles the given schema and sets relevant fields
func (s *Stream) Compile(ctx context.Context, schemaKind StreamSchemaKind, newSchema string, newIndexes *string, description *string, update bool) error {
	// get schema
	var schema schemalang.Schema
	var indexes schemalang.Indexes
	var err error
	switch schemaKind {
	case StreamSchemaKindAvro:
		schema, err = transpilers.FromAvro(newSchema)
	case StreamSchemaKindGraphQL:
		schema, indexes, err = transpilers.FromGraphQL(newSchema)
	default:
		err = fmt.Errorf("Unsupported stream schema definition language '%s'", schemaKind)
	}
	if err != nil {
		return err
	}

	// parse newIndexes if set (overrides output from transpiler)
	if newIndexes != nil {
		err = json.Unmarshal([]byte(*newIndexes), &indexes)
		if err != nil {
			return fmt.Errorf("error parsing indexes: '%s'", err.Error())
		}
	}

	// check schema
	err = schemalang.Check(schema)
	if err != nil {
		return err
	}

	// check indexes
	err = indexes.Check(schema)
	if err != nil {
		return err
	}

	// normalize indexes
	canonicalIndexes := indexes.CanonicalJSON()

	// get avro schemas
	canonicalAvro := transpilers.ToAvro(schema, false)
	avro := transpilers.ToAvro(schema, true)

	// if update, check canonical avro is the same
	if update {
		if canonicalAvro != s.CanonicalAvroSchema {
			return fmt.Errorf("Unfortunately we do not currently support updating a stream's data structure; you can only edit its documentation")
		}
		if canonicalIndexes != s.CanonicalIndexes {
			return fmt.Errorf("Unfortunately we do not currently support updating a stream's indexes")
		}
	}

	// TEMPORARY: disable secondary indexes
	if len(indexes) != 1 {
		return fmt.Errorf("Cannot add secondary indexes to stream (a stream must have exactly one key index)")
	}

	// set indexes
	if !update {
		s.assignIndexes(indexes)
	}

	// set missing stream fields
	s.SchemaKind = schemaKind
	s.Schema = newSchema
	s.AvroSchema = avro
	s.CanonicalAvroSchema = canonicalAvro
	s.CanonicalIndexes = canonicalIndexes
	if description == nil {
		s.Description = schema.(*schemalang.Record).Doc
	} else {
		s.Description = *description
	}

	return nil
}

// Sets StreamIndexes to new StreamIndex objects based on def.
// Doesn't execute any DB actions, so doesn't set any IDs.
func (s *Stream) assignIndexes(indexes schemalang.Indexes) {
	indexes.Sort()
	s.StreamIndexes = make([]*StreamIndex, len(indexes))
	for idx, index := range indexes {
		s.StreamIndexes[idx] = &StreamIndex{
			ShortID:   idx,
			Fields:    index.Fields,
			Primary:   index.Key,
			Normalize: index.Normalize,
		}
	}
}

// Stage creates or updates the stream so that it adheres to the params (or does nothing if everything already conforms).
// If s is a new stream object, Name and ProjectID must be set before calling Stage.
func (s *Stream) Stage(ctx context.Context, schemaKind StreamSchemaKind, schema string, indexes *string, description *string, allowManualWrites *bool, useLog *bool, useIndex *bool, useWarehouse *bool, logRetentionSeconds *int, indexRetentionSeconds *int, warehouseRetentionSeconds *int) error {
	// determine whether to insert or update
	update := (s.StreamID != uuid.Nil)

	// tracks whether a save is necessary
	save := !update

	// compile if necessary
	md5Input := []byte(schema)
	if indexes != nil {
		md5Input = append(md5Input, []byte(*indexes)...)
	}
	schemaMD5 := md5.Sum(md5Input)
	if schemaKind != s.SchemaKind || !bytes.Equal(schemaMD5[:], s.SchemaMD5) {
		err := s.Compile(ctx, schemaKind, schema, indexes, description, update)
		if err != nil {
			return err
		}
		s.SchemaMD5 = schemaMD5[:]
		save = true
	}

	// update fields if necessary

	if allowManualWrites != nil && *allowManualWrites != s.AllowManualWrites {
		s.AllowManualWrites = *allowManualWrites
		save = true
	}

	if update {
		// on updates, don't allow changes to used services and retention

		if useLog != nil && *useLog != s.UseLog {
			return fmt.Errorf("Cannot change useLog after a stream is created")
		}

		if useIndex != nil && *useIndex != s.UseIndex {
			return fmt.Errorf("Cannot change useIndex after a stream is created")
		}

		if useWarehouse != nil && *useWarehouse != s.UseWarehouse {
			return fmt.Errorf("Cannot change useWarehouse after a stream is created")
		}

		if logRetentionSeconds != nil && int32(*logRetentionSeconds) != s.LogRetentionSeconds {
			return fmt.Errorf("Cannot change logRetentionSeconds after a stream is created")
		}

		if indexRetentionSeconds != nil && int32(*indexRetentionSeconds) != s.IndexRetentionSeconds {
			return fmt.Errorf("Cannot change indexRetentionSeconds after a stream is created")
		}

		if warehouseRetentionSeconds != nil && int32(*warehouseRetentionSeconds) != s.WarehouseRetentionSeconds {
			return fmt.Errorf("Cannot change warehouseRetentionSeconds after a stream is created")
		}
	} else {
		save = true

		s.UseLog = true
		if useLog != nil {
			s.UseLog = *useLog
		}

		s.UseIndex = true
		if useIndex != nil {
			s.UseIndex = *useIndex
		}

		s.UseWarehouse = true
		if useWarehouse != nil {
			s.UseWarehouse = *useWarehouse
		}

		if logRetentionSeconds != nil {
			if !s.UseLog {
				return fmt.Errorf("Cannot set logRetentionSeconds on stream that doesn't have useLog=true")
			}
			s.LogRetentionSeconds = int32(*logRetentionSeconds)
		}

		if indexRetentionSeconds != nil {
			if !s.UseIndex {
				return fmt.Errorf("Cannot set indexRetentionSeconds on stream that doesn't have useIndex=true")
			}
			s.IndexRetentionSeconds = int32(*indexRetentionSeconds)
		}

		if warehouseRetentionSeconds != nil {
			if !s.UseWarehouse {
				return fmt.Errorf("Cannot set warehouseRetentionSeconds on stream that doesn't have useWarehouse=true")
			}
			s.WarehouseRetentionSeconds = int32(*warehouseRetentionSeconds)
		}

		// if there's a normalized index, make sure we use a non-expiring log
		for _, index := range s.StreamIndexes {
			if index.Normalize && (!s.UseLog || s.LogRetentionSeconds != 0) {
				return fmt.Errorf("Cannot use normalized indexes when useLog=false or logRetentionSeconds is set")
			}
		}
	}

	// temporarily, we're going to enforce UseLog
	if !s.UseLog {
		return fmt.Errorf("Currently doesn't support streams with useLog=false")
	}

	// quit if no changes
	if !save {
		return nil
	}

	// validate
	err := GetValidator().Struct(s)
	if err != nil {
		return err
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
			instances := FindStreamInstances(tx.Context(), s.StreamID, nil, nil)
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

		return nil
	})
}

// Delete deletes a stream and all its related instances
func (s *Stream) Delete(ctx context.Context) error {
	// delete instances
	instances := FindStreamInstances(ctx, s.StreamID, nil, nil)
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

// StageStreamInstance creates or finds an instance with version in the stream
func (s *Stream) StageStreamInstance(ctx context.Context, version int, makeFinal bool, makePrimary bool) (*StreamInstance, error) {
	if version < 0 {
		return nil, fmt.Errorf("Cannot create stream instance with negative version (got %d)", version)
	}

	// find existing instance
	instances := FindStreamInstances(ctx, s.StreamID, &version, nil)
	maxVersion := 0
	for _, si := range instances {
		// return if version matches
		if si.Version == version {
			// update if makeFinal or makePrimary
			if (makeFinal && si.MadeFinalOn == nil) || (makePrimary && si.MadePrimaryOn == nil) {
				err := s.UpdateStreamInstance(ctx, si, makeFinal, makePrimary)
				if err != nil {
					return nil, err
				}
			}

			// return
			return si, nil
		}

		// track for maxVersion
		if si.Version > maxVersion {
			maxVersion = si.Version
		}
	}

	// don't create if there's an instance with a higher version
	if len(instances) > 0 {
		return nil, fmt.Errorf("One or more instances already exist with a higher version for stream (highest currently has version %d)", maxVersion)
	}

	// check not too many instances
	if s.InstancesCreatedCount-s.InstancesDeletedCount >= MaxInstancesPerStream {
		return nil, gqlerror.Errorf("You cannot have more than %d instances per stream. Delete an existing instance to make room for more.", MaxInstancesPerStream)
	}

	// create instance
	si := &StreamInstance{
		StreamID: s.StreamID,
		Stream:   s,
		Version:  version,
	}

	err := hub.DB.WithContext(ctx).RunInTransaction(func(tx *pg.Tx) error {
		_, err := tx.Model(si).Insert()
		if err != nil {
			return err
		}

		// increment count
		s.InstancesCreatedCount++
		_, err = tx.Model(s).Set("instances_created_count = instances_created_count + 1").WherePK().Update()
		if err != nil {
			return err
		}

		// register instance
		err = hub.Engine.RegisterInstance(tx.Context(), s.Project, s, si)
		if err != nil {
			return err
		}

		// update if makeFinal or makePrimary
		if makeFinal || makePrimary {
			err = s.UpdateStreamInstanceWithTx(tx, si, makeFinal, makePrimary)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return si, nil
}

// UpdateStreamInstance updates the state of a stream instance
func (s *Stream) UpdateStreamInstance(ctx context.Context, instance *StreamInstance, makeFinal bool, makePrimary bool) error {
	return hub.DB.WithContext(ctx).RunInTransaction(func(tx *pg.Tx) error {
		return s.UpdateStreamInstanceWithTx(tx, instance, makeFinal, makePrimary)
	})
}

// UpdateStreamInstanceWithTx is like UpdateStreamInstance, but runs in a transaction
func (s *Stream) UpdateStreamInstanceWithTx(tx *pg.Tx, instance *StreamInstance, makeFinal bool, makePrimary bool) error {
	// check relationship
	if instance.StreamID != s.StreamID {
		return fmt.Errorf("Cannot update instance '%s' because it doesn't belong to stream '%s'", instance.StreamInstanceID.String(), s.StreamID.String())
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

	// delete earlier versions
	if makePrimary {
		instances := FindStreamInstances(tx.Context(), s.StreamID, nil, &instance.Version)
		for _, si := range instances {
			err := s.DeleteStreamInstanceWithTx(tx, si)
			if err != nil {
				return err
			}
			getStreamCache().Clear(tx.Context(), si.StreamInstanceID)
		}
	}

	// clear caches
	getStreamCache().Clear(tx.Context(), instance.StreamInstanceID)
	if makePrimary {
		getInstanceCache().clear(tx.Context(), s.Project.Organization.Name, s.Project.Name, s.Name)
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
