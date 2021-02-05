package stream

import (
	"context"
	"fmt"
	"time"

	uuid "github.com/satori/go.uuid"
	"gitlab.com/beneath-hq/beneath/infra/db"
	"gitlab.com/beneath-hq/beneath/models"
)

// FindStreamInstance finds an instance and related stream details
func (s *Service) FindStreamInstance(ctx context.Context, instanceID uuid.UUID) *models.StreamInstance {
	si := &models.StreamInstance{StreamInstanceID: instanceID}
	err := s.DB.GetDB(ctx).ModelContext(ctx, si).Column(
		"stream_instance.*",
		"Stream",
		"Stream.StreamIndexes",
		"Stream.PrimaryStreamInstance",
		"Stream.Project",
		"Stream.Project.Organization",
	).WherePK().Select()
	if !db.AssertFoundOne(err) {
		return nil
	}
	return si
}

// FindStreamInstanceByOrganizationProjectStreamAndVersion finds an instance and related stream details
func (s *Service) FindStreamInstanceByOrganizationProjectStreamAndVersion(ctx context.Context, organizationName string, projectName string, streamName string, version int) *models.StreamInstance {
	si := &models.StreamInstance{}
	err := s.DB.GetDB(ctx).ModelContext(ctx, si).
		Column(
			"stream_instance.*",
			"Stream",
			"Stream.StreamIndexes",
			"Stream.PrimaryStreamInstance",
			"Stream.Project",
			"Stream.Project.Organization",
		).
		Where("lower(stream__project__organization.name) = lower(?)", organizationName).
		Where("lower(stream__project.name) = lower(?)", projectName).
		Where("lower(stream.name) = lower(?)", streamName).
		Where("stream_instance.version = ?", version).
		Select()
	if !db.AssertFoundOne(err) {
		return nil
	}
	return si
}

// FindStreamInstanceByVersion finds an instance by version in its parent stream
func (s *Service) FindStreamInstanceByVersion(ctx context.Context, streamID uuid.UUID, version int) *models.StreamInstance {
	si := &models.StreamInstance{}
	err := s.DB.GetDB(ctx).ModelContext(ctx, si).
		Column(
			"stream_instance.*",
			"Stream",
			"Stream.StreamIndexes",
			"Stream.PrimaryStreamInstance",
			"Stream.Project",
			"Stream.Project.Organization",
		).
		Where("stream_instance.stream_id = ?", streamID).
		Where("stream_instance.version = ?", version).
		Select()
	if !db.AssertFoundOne(err) {
		return nil
	}
	return si
}

// FindStreamInstances finds the instances associated with the given stream
func (s *Service) FindStreamInstances(ctx context.Context, streamID uuid.UUID, fromVersion *int, toVersion *int) []*models.StreamInstance {
	var res []*models.StreamInstance

	// optimization
	if fromVersion != nil && toVersion != nil && *fromVersion == *toVersion {
		return res
	}

	// build query
	query := s.DB.GetDB(ctx).ModelContext(ctx, &res).Where("stream_id = ?", streamID)
	if fromVersion != nil {
		query = query.Where("version >= ?", *fromVersion)
	}
	if toVersion != nil {
		query = query.Where("version < ?", *toVersion)
	}

	// select
	err := query.Order("version DESC").Select()
	if err != nil {
		panic(fmt.Errorf("error fetching stream instances: %s", err.Error()))
	}

	return res
}

// CreateStreamInstance creates and registers a new stream instance
func (s *Service) CreateStreamInstance(ctx context.Context, stream *models.Stream, version int, makePrimary bool) (*models.StreamInstance, error) {
	if version < 0 {
		return nil, fmt.Errorf("Cannot create stream instance with negative version (got %d)", version)
	}

	instance := &models.StreamInstance{
		StreamID: stream.StreamID,
		Stream:   stream,
		Version:  version,
	}

	if makePrimary {
		now := time.Now()
		instance.MadePrimaryOn = &now
	}

	// create in transaction
	err := s.DB.InTransaction(ctx, func(ctx context.Context) error {
		tx := s.DB.GetDB(ctx)

		_, err := tx.Model(instance).Insert()
		if err != nil {
			return err
		}

		streamQuery := tx.Model(stream).WherePK()
		streamQuery.Set("instances_created_count = instances_created_count + 1")
		if makePrimary {
			streamQuery.Set("primary_stream_instance_id = ?", instance.StreamInstanceID)
			streamQuery.Set("instances_made_primary_count = instances_made_primary_count + 1")
		}
		_, err = streamQuery.Update()
		if err != nil {
			return err
		}

		// modify stream to reflect changes without refetching
		stream.InstancesCreatedCount++
		stream.InstancesMadePrimaryCount++
		stream.PrimaryStreamInstance = instance
		stream.PrimaryStreamInstanceID = &instance.StreamInstanceID

		return nil
	})
	if err != nil {
		return nil, err
	}

	// publish event
	err = s.Bus.Publish(ctx, &models.StreamInstanceCreatedEvent{
		Stream:         stream,
		StreamInstance: instance,
		MakePrimary:    makePrimary,
	})
	if err != nil {
		return nil, err
	}

	return instance, nil
}

// UpdateStreamInstance makes a stream instance primary and/or final
func (s *Service) UpdateStreamInstance(ctx context.Context, stream *models.Stream, instance *models.StreamInstance, makeFinal bool, makePrimary bool) error {
	// makeFinal is one-way and can be done once
	// makePrimary can be done any number of times

	// bail if there's nothing to do
	if (!makeFinal || instance.MadeFinalOn != nil) && !makePrimary {
		return nil
	}

	// run in transaction
	now := time.Now()
	err := s.DB.InTransaction(ctx, func(ctx context.Context) error {
		tx := s.DB.GetDB(ctx)
		updateStreamQuery := tx.Model(stream).WherePK()
		updateInstanceQuery := tx.Model(instance).WherePK()

		// handle makeFinal
		if makeFinal && instance.MadeFinalOn == nil {
			updateStreamQuery.Set("instances_made_final_count = instances_made_final_count + 1")
			updateInstanceQuery.Set("made_final_on = now()")
			instance.MadeFinalOn = &now
		}

		// handle makePrimary
		if makePrimary {
			updateInstanceQuery.Set("made_primary_on = now()")
			instance.MadePrimaryOn = &now
			updateStreamQuery.Set("primary_stream_instance_id = ?", instance.StreamInstanceID)
			if instance.MadePrimaryOn == nil {
				updateStreamQuery.Set("instances_made_primary_count = instances_made_primary_count + 1")
			}
		}

		// updated on timestamps
		updateStreamQuery.Set("updated_on = now()")
		updateInstanceQuery.Set("updated_on = now()")
		instance.UpdatedOn = now

		// update
		_, err := updateStreamQuery.Update()
		if err != nil {
			return err
		}
		_, err = updateInstanceQuery.Update()
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	// publish event
	err = s.Bus.Publish(ctx, &models.StreamInstanceUpdatedEvent{
		Stream:           stream,
		StreamInstance:   instance,
		MakeFinal:        makeFinal,
		MakePrimary:      makePrimary,
		OrganizationName: stream.Project.Organization.Name,
		ProjectName:      stream.Project.Name,
	})
	if err != nil {
		return err
	}

	return nil
}

// DeleteStreamInstance deletes a stream and its data
func (s *Service) DeleteStreamInstance(ctx context.Context, stream *models.Stream, instance *models.StreamInstance) error {
	err := s.DB.InTransaction(ctx, func(ctx context.Context) error {
		tx := s.DB.GetDB(ctx)

		err := tx.Delete(instance)
		if err != nil {
			return err
		}

		_, err = tx.Model(stream).
			WherePK().
			Set("updated_on = ?", stream.UpdatedOn).
			Set("instances_deleted_count = instances_deleted_count + 1").
			Update()
		if err != nil {
			return err
		}

		// stream.PrimaryStreamInstanceID has SET NULL constraint (so we don't have to handle it)

		return nil
	})
	if err != nil {
		return err
	}

	// publish event
	err = s.Bus.Publish(ctx, &models.StreamInstanceDeletedEvent{
		Stream:         stream,
		StreamInstance: instance,
	})
	if err != nil {
		return err
	}

	return nil
}
