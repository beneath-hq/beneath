package entity

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"time"

	"gitlab.com/beneath-hq/beneath/pkg/codec"

	"github.com/go-pg/pg/v9"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/go-playground/validator.v9"

	"gitlab.com/beneath-hq/beneath/control/taskqueue"
	"gitlab.com/beneath-hq/beneath/internal/hub"
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

	// Schema-related fields (note: used in stream cache)
	Schema              string `sql:",notnull",validate:"required"`
	AvroSchema          string `sql:",type:json,notnull",validate:"required"`
	CanonicalAvroSchema string `sql:",type:json,notnull",validate:"required"`
	BigQuerySchema      string `sql:"bigquery_schema,type:json,notnull",validate:"required"`

	// Indexes (note: used in stream cache)
	StreamIndexes []*StreamIndex

	// Behaviour-related fields (note: used in stream cache)
	External         bool  `sql:",notnull"`
	Batch            bool  `sql:",notnull"`
	Manual           bool  `sql:",notnull"`
	RetentionSeconds int32 `sql:",notnull,default:0"`

	// Instances-related fields
	InstancesCreatedCount   int32 `sql:",notnull,default:0"`
	InstancesCommittedCount int32 `sql:",notnull,default:0"`
	StreamInstances         []*StreamInstance
	CurrentStreamInstanceID *uuid.UUID `sql:"on_delete:SET NULL,type:uuid"`
	CurrentStreamInstance   *StreamInstance
}

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
		Column("stream.*", "Project", "Project.Organization", "CurrentStreamInstance", "SourceModel", "StreamIndexes").
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
		Column("stream.*", "Project", "Project.Organization", "CurrentStreamInstance", "SourceModel", "StreamIndexes"). // Q: should I not select the whole organization table? and instead just select the name?
		Where("lower(project__organization.name) = lower(?)", organizationName).
		Where("lower(project.name) = lower(?)", projectName).
		Where("lower(stream.name) = lower(?)", streamName).
		Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return stream
}

// FindInstanceIDByOrganizationProjectAndName returns the current instance ID of the stream
func FindInstanceIDByOrganizationProjectAndName(ctx context.Context, organizationName string, projectName string, streamName string) uuid.UUID {
	return getInstanceCache().get(ctx, organizationName, projectName, streamName)
}

// FindCachedStreamByCurrentInstanceID returns select info about the instance's stream
func FindCachedStreamByCurrentInstanceID(ctx context.Context, instanceID uuid.UUID) *CachedStream {
	return getStreamCache().Get(ctx, instanceID)
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

		if !s.indexesEqual(streamDef) {
			return fmt.Errorf("Cannot change stream key or indexes in an update")
		}
	}

	// set indexes
	if !update {
		s.assignIndexes(streamDef)
	}

	// set missing stream fields
	s.Name = streamDef.Name
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

	// populate s.Project.Organization if not set
	if s.Project.Organization == nil {
		s.Project.Organization = FindOrganization(ctx, s.Project.OrganizationID)
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

// CreateWithTx creates the stream and an associated instance (if streaming) using tx
func (s *Stream) CreateWithTx(tx *pg.Tx) error {
	// insert stream
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
	s.UpdatedOn = time.Now()
	_, err := tx.Model(s).WherePK().Update()
	if err != nil {
		return err
	}

	// get instances
	if len(s.StreamInstances) == 0 {
		err := tx.Model((*StreamInstance)(nil)).
			Where("stream_id = ?", s.StreamID).
			Select(&s.StreamInstances)
		if err != nil {
			return err
		}
	}

	// Clear stream cache
	for _, si := range s.StreamInstances {
		getStreamCache().Clear(tx.Context(), si.StreamInstanceID)
	}

	// update in bigquery
	if s.CurrentStreamInstanceID != nil {
		cs := FindCachedStreamByCurrentInstanceID(tx.Context(), *s.CurrentStreamInstanceID)
		err = hub.Engine.RegisterInstance(tx.Context(), cs, cs, cs)
		if err != nil {
			return err
		}
	}

	return nil
}

// Delete deletes a stream and all its related instances
func (s *Stream) Delete(ctx context.Context) error {
	// get instances
	var instances []*StreamInstance
	err := hub.DB.ModelContext(ctx, &instances).
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
	err = hub.DB.WithContext(ctx).Delete(s)
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

// CompileAndCreate compiles the schema, derives name and avro schemas and inserts
// the stream into the database
func (s *Stream) CompileAndCreate(ctx context.Context) error {
	// compile
	err := s.Compile(ctx, false)
	if err != nil {
		return err
	}

	// create stream (and a new stream instance ID if not batch)
	err = hub.DB.WithContext(ctx).RunInTransaction(func(tx *pg.Tx) error {
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

	return hub.DB.WithContext(ctx).RunInTransaction(func(tx *pg.Tx) error {
		return s.UpdateWithTx(tx)
	})
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
		return nil, fmt.Errorf("Another batch is already outstanding for stream '%s/%s/%s' – commit or clear it before continuing", s.Project.Organization.Name, s.Project.Name, s.Name)
	}

	// create new
	si := &StreamInstance{StreamID: s.StreamID}
	_, err = tx.Model(si).Insert()
	if err != nil {
		return nil, err
	}
	si.Stream = s

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

// CommitStreamInstance promotes an instance to current_instance_id and deletes the old instance
func (s *Stream) CommitStreamInstance(ctx context.Context, instance *StreamInstance) error {
	return hub.DB.WithContext(ctx).RunInTransaction(func(tx *pg.Tx) error {
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
	s.UpdatedOn = time.Now()
	_, err := tx.Model(s).Column("current_stream_instance_id", "updated_on").WherePK().Update()
	if err != nil {
		return err
	}

	// set committed_on
	instance.CommittedOn = time.Now()
	_, err = tx.Model(instance).Column("committed_on").WherePK().Update()
	if err != nil {
		return err
	}

	// clear instance cache in redis
	getInstanceCache().clear(tx.Context(), s.Project.Organization.Name, s.Project.Name, s.Name)
	getStreamCache().Clear(tx.Context(), instance.StreamInstanceID)
	if prevInstanceID != nil {
		getStreamCache().Clear(tx.Context(), *prevInstanceID)
	}

	// increment count
	s.InstancesCommittedCount++
	_, err = tx.Model(s).Set("instances_committed_count = instances_committed_count + 1").WherePK().Update()
	if err != nil {
		return err
	}

	// call on warehouse
	err = hub.Engine.PromoteInstance(tx.Context(), s.Project, s, instance)
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
	return hub.DB.WithContext(ctx).RunInTransaction(func(tx *pg.Tx) error {
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
		s.UpdatedOn = time.Now()
		_, err := tx.Model(s).Column("current_stream_instance_id", "updated_on").WherePK().Update()
		if err != nil {
			return err
		}
		getInstanceCache().clear(tx.Context(), s.Project.Organization.Name, s.Project.Name, s.Name)
	}

	// delete
	err := tx.Delete(si)
	if err != nil {
		return err
	}

	// deregister
	err = taskqueue.Submit(tx.Context(), &CleanupInstanceTask{
		CachedStream: NewCachedStream(s, si.StreamInstanceID),
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
		err := hub.DB.ModelContext(ctx, (*StreamInstance)(nil)).
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
