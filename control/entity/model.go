package entity

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/go-playground/validator.v9"

	"gitlab.com/beneath-hq/beneath/internal/hub"
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
	ProjectID     uuid.UUID `sql:"on_delete:RESTRICT,notnull,type:uuid"`
	Project       *Project
	InputStreams  []*Stream `pg:"many2many:streams_into_models,fk:model_id,joinFK:stream_id"`
	OutputStreams []*Stream `pg:"fk:source_model_id"`
	ServiceID     uuid.UUID `sql:"on_delete:RESTRICT,notnull,type:uuid"`
	Service       *Service
}

// StreamIntoModel represnts the many-to-many relationship between input streams and models
type StreamIntoModel struct {
	tableName struct{}  `sql:"streams_into_models,alias:sm"`
	StreamID  uuid.UUID `sql:"on_delete:CASCADE,pk,type:uuid"`
	Stream    *Stream
	ModelID   uuid.UUID `sql:"on_delete:CASCADE,pk,type:uuid"`
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

var (
	// used for validation
	modelNameRegex *regexp.Regexp
)

func init() {
	modelNameRegex = regexp.MustCompile("^[_a-z][_a-z0-9]*$")
	orm.RegisterTable((*StreamIntoModel)(nil))
	GetValidator().RegisterStructValidation(modelValidation, Model{})
}

// custom model validation
func modelValidation(sl validator.StructLevel) {
	m := sl.Current().Interface().(Model)
	if !modelNameRegex.MatchString(m.Name) {
		sl.ReportError(m.Name, "Name", "", "alphanumericorunderscore", "")
	}
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
	err := hub.DB.ModelContext(ctx, model).
		Column("model.*").
		Relation("Project").
		Relation("OutputStreams").
		Relation("InputStreams").
		Relation("Service").
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
	err := hub.DB.ModelContext(ctx, model).
		Column("model.*").
		Relation("Project").
		Relation("OutputStreams").
		Relation("InputStreams").
		Relation("Service").
		Where("lower(model.name) = lower(?)", name).
		Where("lower(project.name) = lower(?)", projectName).
		Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return model
}

// CompileAndCreate creates the model and its dependencies
func (m *Model) CompileAndCreate(ctx context.Context, inputStreamIDs []uuid.UUID, outputStreamScheams []string, readQuota int64, writeQuota int64) error {
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
		stream, err := m.compileSchema(ctx, schema)
		if err != nil {
			return err
		}
		outputStreams[idx] = stream
	}

	// validate model
	err := GetValidator().Struct(m)
	if err != nil {
		return err
	}

	// begin database transaction
	err = hub.DB.WithContext(ctx).RunInTransaction(func(tx *pg.Tx) error {
		// insert model
		_, err = tx.Model(m).Insert()
		if err != nil {
			return err
		}

		// create service
		service := &Service{
			Name:           fmt.Sprintf("Model %s (%s)", m.ModelID.String(), m.Name),
			Kind:           ServiceKindModel,
			OrganizationID: m.Project.OrganizationID,
			ReadQuota:      readQuota,
			WriteQuota:     writeQuota,
		}
		_, err := tx.Model(service).Insert()
		if err != nil {
			return err
		}

		// update model with service ID
		m.ServiceID = service.ServiceID
		_, err = tx.Model(m).Column("service_id").WherePK().Update()
		if err != nil {
			return err
		}

		// insert StreamIntoModels and PermissionsServicesStreams
		for _, streamID := range inputStreamIDs {
			_, err = tx.Model(&StreamIntoModel{
				ModelID:  m.ModelID,
				StreamID: streamID,
			}).Insert()
			if err != nil {
				return err
			}

			_, err = tx.Model(&PermissionsServicesStreams{
				ServiceID: m.ServiceID,
				StreamID:  streamID,
				Read:      true,
				Write:     false,
			}).Insert()
			if err != nil {
				return err
			}
		}

		// insert outputStreams and PermissionsServicesStreams
		for _, stream := range outputStreams {
			stream.SourceModelID = &m.ModelID
			err := stream.CreateWithTx(tx)
			if err != nil {
				return err
			}

			_, err = tx.Model(&PermissionsServicesStreams{
				ServiceID: m.ServiceID,
				StreamID:  stream.StreamID,
				Read:      true,
				Write:     true,
			}).Insert()
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

// CompileAndUpdate ...
func (m *Model) CompileAndUpdate(ctx context.Context, inputStreamIDs []uuid.UUID, outputStreamScheams []string, readQuota int64, writeQuota int64) error {
	// check inputStreamIDs contain all current input streams
	for _, s := range m.InputStreams {
		if !m.containsUUID(inputStreamIDs, s.StreamID) {
			return fmt.Errorf("Cannot delete input streams from existing model")
		}
	}

	// find new input streams
	var newInputStreamIDs []uuid.UUID
	for _, id := range inputStreamIDs {
		// search for ID in InputStreams
		found := false
		for _, s := range m.InputStreams {
			if s.StreamID == id {
				found = true
			}
		}

		// if not found, add to newInputStreamIDs
		if !found {
			newInputStreamIDs = append(newInputStreamIDs, id)
		}
	}

	// update existing output stream schemas
	var updateOutputStreams []*Stream
	outputNames := make(map[string]bool)
	for _, outputSchema := range outputStreamScheams {
		// compile new to get information about it
		new, err := m.compileSchema(ctx, outputSchema)
		if err != nil {
			return err
		}

		// check not multiple schemas with same name
		if outputNames[new.Name] {
			return fmt.Errorf("Found two output schemas with the name '%s'. "+"Stream names must be unique.", new.Name)
		}
		outputNames[new.Name] = true

		// find matching existing stream
		var existing *Stream
		for _, s := range m.OutputStreams {
			if s.Name == new.Name {
				existing = s
			}
		}
		if existing == nil {
			return fmt.Errorf("Couldn't find existing output stream with name '%s'. "+
				"You can only update existing streams, not add new streams.", new.Name)
		}

		// make sure new input streams aren't an existing output (weak cycle detection)
		if m.containsUUID(newInputStreamIDs, existing.StreamID) {
			return fmt.Errorf("You cannot use the output stream '%s' as input to this model (cycle)", existing.Name)
		}

		// skip if no changes
		if existing.Schema == outputSchema {
			continue
		}

		// update
		existing.ProjectID = m.ProjectID
		existing.Schema = outputSchema
		err = existing.Compile(ctx, true)
		if err != nil {
			return err
		}

		// add for saving
		updateOutputStreams = append(updateOutputStreams, existing)
	}

	// validate model
	err := GetValidator().Struct(m)
	if err != nil {
		return err
	}

	// update
	return hub.DB.WithContext(ctx).RunInTransaction(func(tx *pg.Tx) error {
		// update model
		m.UpdatedOn = time.Now()
		_, err := tx.Model(m).WherePK().Update()
		if err != nil {
			return err
		}

		// update quotas
		m.Service.ReadQuota = readQuota
		m.Service.WriteQuota = writeQuota
		_, err = tx.Model(m.Service).Column("read_quota", "write_quota").WherePK().Update()
		if err != nil {
			return err
		}

		// insert StreamIntoModels and PermissionsServicesStreams
		for _, streamID := range newInputStreamIDs {
			_, err = tx.Model(&StreamIntoModel{
				ModelID:  m.ModelID,
				StreamID: streamID,
			}).Insert()
			if err != nil {
				return err
			}

			_, err = tx.Model(&PermissionsServicesStreams{
				ServiceID: m.ServiceID,
				StreamID:  streamID,
				Read:      true,
				Write:     false,
			}).Insert()
			if err != nil {
				return err
			}
		}

		// insert outputStreams
		for _, stream := range updateOutputStreams {
			err := stream.UpdateWithTx(tx)
			if err != nil {
				return err
			}
		}

		// done
		return nil
	})
}

func (m *Model) compileSchema(ctx context.Context, schema string) (*Stream, error) {
	stream := &Stream{
		Schema:    schema,
		External:  false,
		Batch:     m.Kind == ModelKindBatch,
		Manual:    false,
		ProjectID: m.ProjectID,
		Project:   m.Project,
	}
	err := stream.Compile(ctx, false)
	if err != nil {
		return nil, err
	}
	return stream, nil
}

func (m *Model) containsUUID(haystack []uuid.UUID, needle uuid.UUID) bool {
	for _, id := range haystack {
		if id == needle {
			return true
		}
	}
	return false
}

// Delete deletes the model and it's dependent streams
func (m *Model) Delete(ctx context.Context) error {
	if m.OutputStreams == nil {
		m = FindModel(ctx, m.ModelID)
		if m == nil {
			return nil
		}
	}

	for _, stream := range m.OutputStreams {
		if stream.Project == nil {
			// faster than fetching individually
			stream.Project = m.Project
		}

		err := stream.Delete(ctx)
		if err != nil {
			return err
		}
	}

	return hub.DB.WithContext(ctx).RunInTransaction(func(tx *pg.Tx) error {
		err := tx.Delete(m)
		if err != nil {
			return err
		}

		err = tx.Delete(&Service{ServiceID: m.ServiceID})
		if err != nil {
			return err
		}

		return nil
	})
}

// CreateBatch prepares a new batch of stream instances for OutputStreams
func (m *Model) CreateBatch(ctx context.Context) ([]*StreamInstance, error) {
	instances := make([]*StreamInstance, len(m.OutputStreams))
	for i, stream := range m.OutputStreams {
		stream.Project = m.Project
		si, err := stream.CreateStreamInstance(ctx)
		if err != nil {
			return nil, err
		}
		instances[i] = si
	}
	return instances, nil
}

// CommitBatch commits a batch created with CreateBatch
func (m *Model) CommitBatch(ctx context.Context, instanceIDs []uuid.UUID) error {
	// get instances
	var instances []*StreamInstance
	err := hub.DB.ModelContext(ctx, &instances).
		Where("stream_instance_id in (?)", pg.In(instanceIDs)).
		Relation("Stream", func(q *orm.Query) (*orm.Query, error) {
			return q.Where("source_model_id = ?", m.ModelID), nil
		}).
		Select()
	if err != nil {
		return err
	}

	// We want to guard against being supplied instance IDs that don't belong to streams of this model.
	// And also against being supplied bad instance IDs or fewer instance IDs than there are output streams.
	// These checks should guard againts all that.
	if len(instances) != len(instanceIDs) || len(instances) != len(m.OutputStreams) {
		return fmt.Errorf("found one or more unexpected instance IDs in commit")
	}
	for _, instance := range instances {
		if instance.Stream == nil {
			return fmt.Errorf("found instance that don't belong to stream of this model")
		}
	}

	// commit
	for _, instance := range instances {
		instance.Stream.Project = m.Project
		err := instance.Stream.CommitStreamInstance(ctx, instance)
		if err != nil {
			return err
		}
	}
	return nil
}

// ClearPendingBatches clears all instance IDs not promoted to current
func (m *Model) ClearPendingBatches(ctx context.Context) error {
	for _, stream := range m.OutputStreams {
		stream.Project = m.Project
		err := stream.ClearPendingBatches(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}
