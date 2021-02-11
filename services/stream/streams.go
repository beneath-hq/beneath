package stream

import (
	"bytes"
	"context"
	"fmt"
	"time"

	uuid "github.com/satori/go.uuid"
	"gitlab.com/beneath-hq/beneath/infra/db"
	"gitlab.com/beneath-hq/beneath/models"
)

// FindStream finds a stream
func (s *Service) FindStream(ctx context.Context, streamID uuid.UUID) *models.Stream {
	stream := &models.Stream{StreamID: streamID}
	err := s.DB.GetDB(ctx).ModelContext(ctx, stream).
		WherePK().
		Column(
			"stream.*",
			"Project",
			"Project.Organization",
			"PrimaryStreamInstance",
			"StreamIndexes",
		).
		Select()
	if !db.AssertFoundOne(err) {
		return nil
	}
	return stream
}

// FindStreamByOrganizationProjectAndName finds a stream
func (s *Service) FindStreamByOrganizationProjectAndName(ctx context.Context, organizationName string, projectName string, streamName string) *models.Stream {
	stream := &models.Stream{}
	err := s.DB.GetDB(ctx).ModelContext(ctx, stream).
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
	if !db.AssertFoundOne(err) {
		return nil
	}
	return stream
}

// FindStreamsForUser finds the streams that the user has access to
func (s *Service) FindStreamsForUser(ctx context.Context, userID uuid.UUID) []*models.Stream {
	var streams []*models.Stream
	err := s.DB.GetDB(ctx).ModelContext(ctx, &streams).
		Column("stream.*", "Project.project_id", "Project.name", "Project.Organization.organization_id", "Project.Organization.name").
		Join("JOIN permissions_users_projects AS pup ON pup.project_id = stream.project_id").
		Where("pup.user_id = ?", userID).
		Order("project__organization.name", "project.name", "stream.name").
		Limit(200).
		Select()
	if err != nil {
		panic(err)
	}
	return streams
}

// CreateStream compiles and creates a new stream
func (s *Service) CreateStream(ctx context.Context, msg *models.CreateStreamCommand) (*models.Stream, error) {
	stream := &models.Stream{
		Name:      msg.Name,
		Project:   msg.Project,
		ProjectID: msg.Project.ProjectID,
	}

	// compile stream
	err := s.CompileToStream(stream, msg.SchemaKind, msg.Schema, msg.Indexes, msg.Description)
	if err != nil {
		return nil, err
	}
	stream.SchemaMD5 = s.ComputeSchemaMD5(msg.Schema, msg.Indexes)

	// set other values
	stream.Meta = derefBool(msg.Meta, false)
	stream.AllowManualWrites = derefBool(msg.AllowManualWrites, false)
	stream.UseLog = derefBool(msg.UseLog, true)
	stream.UseIndex = derefBool(msg.UseIndex, true)
	stream.UseWarehouse = derefBool(msg.UseWarehouse, true)

	if msg.LogRetentionSeconds != nil {
		if !stream.UseLog {
			return nil, fmt.Errorf("Cannot set logRetentionSeconds on stream that doesn't have useLog=true")
		}
		stream.LogRetentionSeconds = int32(*msg.LogRetentionSeconds)
	}

	if msg.IndexRetentionSeconds != nil {
		if !stream.UseIndex {
			return nil, fmt.Errorf("Cannot set indexRetentionSeconds on stream that doesn't have useIndex=true")
		}
		stream.IndexRetentionSeconds = int32(*msg.IndexRetentionSeconds)
	}

	if msg.WarehouseRetentionSeconds != nil {
		if !stream.UseWarehouse {
			return nil, fmt.Errorf("Cannot set warehouseRetentionSeconds on stream that doesn't have useWarehouse=true")
		}
		stream.WarehouseRetentionSeconds = int32(*msg.WarehouseRetentionSeconds)
	}

	// if there's a normalized index, make sure we use a non-expiring log.
	// note that this is currently a little fictional, as normalized index support is pretty shaky.
	for _, index := range stream.StreamIndexes {
		if index.Normalize && (!stream.UseLog || stream.LogRetentionSeconds != 0) {
			return nil, fmt.Errorf("Cannot use normalized indexes when useLog=false or logRetentionSeconds is set")
		}
	}

	// temporarily, we're going to enforce UseLog
	if !stream.UseLog {
		return nil, fmt.Errorf("Currently doesn't support streams with useLog=false")
	}

	// validate
	err = stream.Validate()
	if err != nil {
		return nil, err
	}

	// insert stream and its indexes transactionally
	err = s.DB.InTransaction(ctx, func(ctx context.Context) error {
		tx := s.DB.GetDB(ctx)

		_, err := tx.Model(stream).Insert()
		if err != nil {
			return err
		}

		for _, index := range stream.StreamIndexes {
			index.StreamID = stream.StreamID
			_, err := tx.Model(index).Insert()
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// publish events
	err = s.Bus.Publish(ctx, &models.StreamCreatedEvent{
		Stream: stream,
	})
	if err != nil {
		return nil, err
	}

	return stream, nil
}

// UpdateStream updates a streams schema (within limits) and more
func (s *Service) UpdateStream(ctx context.Context, msg *models.UpdateStreamCommand) error {
	stream := msg.Stream
	updateFields := []string{}
	modifiedSchema := false

	// re-compile and set updates if schema changed
	if msg.SchemaKind != nil && msg.Schema != nil {
		schemaMD5 := s.ComputeSchemaMD5(*msg.Schema, msg.Indexes)
		if !bytes.Equal(schemaMD5, stream.SchemaMD5) {
			err := s.CompileToStream(stream, *msg.SchemaKind, *msg.Schema, &stream.CanonicalIndexes, msg.Description)
			if err != nil {
				return err
			}
			stream.SchemaMD5 = schemaMD5
			updateFields = append(
				updateFields,
				"schema_kind",
				"schema",
				"schema_md5",
				"avro_schema",
				"canonical_avro_schema",
				"canonical_indexes",
				"description",
			)
			modifiedSchema = true
		}
	}

	// update description (you can change a description without changing the schema)
	// (this steps on the former clause's toes, but the if-check here makes them mutually exclusive)
	if msg.Description != nil && *msg.Description != stream.Description {
		stream.Description = *msg.Description
		updateFields = append(updateFields, "description")
	}

	// update meta
	if msg.Meta != nil && *msg.Meta != stream.Meta {
		stream.Meta = *msg.Meta
		updateFields = append(updateFields, "meta")
	}

	// update manual writes
	if msg.AllowManualWrites != nil && *msg.AllowManualWrites != stream.AllowManualWrites {
		stream.AllowManualWrites = *msg.AllowManualWrites
		updateFields = append(updateFields, "allow_manual_writes")
	}

	// quit if no changes
	if len(updateFields) == 0 {
		return nil
	}

	// validate
	err := stream.Validate()
	if err != nil {
		return err
	}

	// update
	stream.UpdatedOn = time.Now()
	_, err = s.DB.GetDB(ctx).ModelContext(ctx, stream).WherePK().Update()
	if err != nil {
		return err
	}

	// publish events
	err = s.Bus.Publish(ctx, &models.StreamUpdatedEvent{
		Stream:         stream,
		ModifiedSchema: modifiedSchema,
	})
	if err != nil {
		return err
	}

	return nil
}

// DeleteStream deletes a stream and all its related instances
func (s *Service) DeleteStream(ctx context.Context, stream *models.Stream) error {
	// delete instances
	instances := s.FindStreamInstances(ctx, stream.StreamID, nil, nil)
	for _, inst := range instances {
		err := s.DeleteStreamInstance(ctx, stream, inst)
		if err != nil {
			return err
		}
	}

	// delete stream
	_, err := s.DB.GetDB(ctx).ModelContext(ctx, stream).WherePK().Delete()
	if err != nil {
		return err
	}

	// publish event
	err = s.Bus.Publish(ctx, &models.StreamDeletedEvent{StreamID: stream.StreamID})
	if err != nil {
		return err
	}

	return nil
}
