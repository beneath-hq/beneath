package resolver

import (
	"context"

	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"gitlab.com/beneath-hq/beneath/models"
	"gitlab.com/beneath-hq/beneath/server/control/gql"
	"gitlab.com/beneath-hq/beneath/services/middleware"
)

// Stream returns the gql.StreamResolver
func (r *Resolver) Stream() gql.StreamResolver {
	return &streamResolver{r}
}

type streamResolver struct{ *Resolver }

func (r *streamResolver) StreamID(ctx context.Context, obj *models.Stream) (string, error) {
	return obj.StreamID.String(), nil
}

func (r *queryResolver) StreamByID(ctx context.Context, streamID uuid.UUID) (*models.Stream, error) {
	stream := r.Streams.FindStream(ctx, streamID)
	if stream == nil {
		return nil, gqlerror.Errorf("Stream with ID %s not found", streamID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.StreamPermissionsForSecret(ctx, secret, stream.StreamID, stream.ProjectID, stream.Project.Public)
	if !perms.Read {
		return nil, gqlerror.Errorf("Not allowed to read stream with ID %s", streamID.String())
	}

	return stream, nil
}

func (r *queryResolver) StreamByOrganizationProjectAndName(ctx context.Context, organizationName string, projectName string, streamName string) (*models.Stream, error) {
	stream := r.Streams.FindStreamByOrganizationProjectAndName(ctx, organizationName, projectName, streamName)
	if stream == nil {
		return nil, gqlerror.Errorf("Stream %s/%s/%s not found", organizationName, projectName, streamName)
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.StreamPermissionsForSecret(ctx, secret, stream.StreamID, stream.ProjectID, stream.Project.Public)
	if !perms.Read && !perms.Write {
		return nil, gqlerror.Errorf("Not allowed to find stream %s/%s/%s", organizationName, projectName, streamName)
	}

	return stream, nil
}

func (r *queryResolver) StreamInstancesForStream(ctx context.Context, streamID uuid.UUID) ([]*models.StreamInstance, error) {
	stream := r.Streams.FindStream(ctx, streamID)
	if stream == nil {
		return nil, gqlerror.Errorf("Stream with ID %s not found", streamID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.StreamPermissionsForSecret(ctx, secret, stream.StreamID, stream.ProjectID, stream.Project.Public)
	if !perms.Read {
		return nil, gqlerror.Errorf("Not allowed to read stream with ID %s", streamID.String())
	}

	instances := r.Streams.FindStreamInstances(ctx, streamID, nil, nil)
	return instances, nil
}

func (r *queryResolver) StreamInstancesByOrganizationProjectAndStreamName(ctx context.Context, organizationName string, projectName string, streamName string) ([]*models.StreamInstance, error) {
	stream := r.Streams.FindStreamByOrganizationProjectAndName(ctx, organizationName, projectName, streamName)
	if stream == nil {
		return nil, gqlerror.Errorf("Stream %s/%s/%s not found", organizationName, projectName, streamName)
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.StreamPermissionsForSecret(ctx, secret, stream.StreamID, stream.ProjectID, stream.Project.Public)
	if !perms.Read {
		return nil, gqlerror.Errorf("Not allowed to read stream with ID %s", stream.StreamID.String())
	}

	instances := r.Streams.FindStreamInstances(ctx, stream.StreamID, nil, nil)
	return instances, nil
}

func (r *queryResolver) StreamsForUser(ctx context.Context, userID uuid.UUID) ([]*models.Stream, error) {
	secret := middleware.GetSecret(ctx)
	if !(secret.IsUser() && secret.GetOwnerID() == userID) {
		return nil, gqlerror.Errorf("StreamsForUser can only be called for the calling user")
	}
	return r.Streams.FindStreamsForUser(ctx, userID), nil
}

func (r *queryResolver) CompileSchema(ctx context.Context, input gql.CompileSchemaInput) (*gql.CompileSchemaOutput, error) {
	stream := &models.Stream{}
	err := r.Streams.CompileToStream(stream, input.SchemaKind, input.Schema, input.Indexes, nil)
	if err != nil {
		return nil, gqlerror.Errorf("Error compiling schema: %s", err.Error())
	}

	canonicalIndexes := &gql.CompileSchemaOutput{
		CanonicalIndexes: stream.CanonicalIndexes,
	}

	return canonicalIndexes, nil
}

func (r *mutationResolver) StageStream(ctx context.Context, organizationName string, projectName string, streamName string, schemaKind models.StreamSchemaKind, schema string, indexes *string, description *string, allowManualWrites *bool, useLog *bool, useIndex *bool, useWarehouse *bool, logRetentionSeconds *int, indexRetentionSeconds *int, warehouseRetentionSeconds *int) (*models.Stream, error) {
	var project *models.Project
	var stream *models.Stream

	secret := middleware.GetSecret(ctx)

	stream = r.Streams.FindStreamByOrganizationProjectAndName(ctx, organizationName, projectName, streamName)
	if stream != nil {
		perms := r.Permissions.ProjectPermissionsForSecret(ctx, secret, stream.ProjectID, false)
		if !perms.Create {
			return nil, gqlerror.Errorf("Not allowed to create or modify resources in project %s/%s", organizationName, projectName)
		}

		err := r.Streams.UpdateStream(ctx, &models.UpdateStreamCommand{
			Stream:            stream,
			SchemaKind:        &schemaKind,
			Schema:            &schema,
			Description:       description,
			AllowManualWrites: allowManualWrites,
		})
		if err != nil {
			return nil, gqlerror.Errorf("Error updating stream: %s", err.Error())
		}
		return stream, nil
	}

	project = r.Projects.FindProjectByOrganizationAndName(ctx, organizationName, projectName)
	if project == nil {
		return nil, gqlerror.Errorf("Project %s/%s not found", organizationName, projectName)
	}

	perms := r.Permissions.ProjectPermissionsForSecret(ctx, secret, project.ProjectID, project.Public)
	if !perms.Create {
		return nil, gqlerror.Errorf("Not allowed to create or modify resources in project %s/%s", organizationName, projectName)
	}

	stream, err := r.Streams.CreateStream(ctx, &models.CreateStreamCommand{
		Project:                   project,
		Name:                      streamName,
		SchemaKind:                schemaKind,
		Schema:                    schema,
		Indexes:                   indexes,
		Description:               description,
		AllowManualWrites:         allowManualWrites,
		UseLog:                    useLog,
		UseIndex:                  useIndex,
		UseWarehouse:              useWarehouse,
		LogRetentionSeconds:       logRetentionSeconds,
		IndexRetentionSeconds:     indexRetentionSeconds,
		WarehouseRetentionSeconds: warehouseRetentionSeconds,
	})
	if err != nil {
		return nil, gqlerror.Errorf("Error creating stream: %s", err.Error())
	}

	_, err = r.Streams.CreateStreamInstance(ctx, stream, 0, true)
	if err != nil {
		return nil, gqlerror.Errorf("Error creating first instance: %s", err.Error())
	}

	return stream, nil
}

func (r *mutationResolver) DeleteStream(ctx context.Context, streamID uuid.UUID) (bool, error) {
	stream := r.Streams.FindStream(ctx, streamID)
	if stream == nil {
		return false, gqlerror.Errorf("Stream %s not found", streamID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.ProjectPermissionsForSecret(ctx, secret, stream.ProjectID, false)
	if !perms.Create {
		return false, gqlerror.Errorf("Not allowed to perform admin functions in project %s", stream.Project.Name)
	}

	err := r.Streams.DeleteStream(ctx, stream)
	if err != nil {
		return false, gqlerror.Errorf("%s", err.Error())
	}

	return true, nil
}

// MaxInstancesPerStream sets a limit for the number of instances for a stream at any given time
const MaxInstancesPerStream = 25

func (r *mutationResolver) StageStreamInstance(ctx context.Context, streamID uuid.UUID, version int, makeFinal *bool, makePrimary *bool) (*models.StreamInstance, error) {
	falseVal := false
	if makeFinal == nil {
		makeFinal = &falseVal
	}
	if makePrimary == nil {
		makePrimary = &falseVal
	}

	secret := middleware.GetSecret(ctx)

	instance := r.Streams.FindStreamInstanceByVersion(ctx, streamID, version)
	if instance != nil {
		perms := r.Permissions.StreamPermissionsForSecret(ctx, secret, instance.Stream.StreamID, instance.Stream.ProjectID, false)
		if !perms.Write {
			return nil, gqlerror.Errorf("Not allowed to write to stream %s/%s", instance.Stream.Name, instance.Stream.Project.Name)
		}

		err := r.Streams.UpdateStreamInstance(ctx, instance.Stream, instance, *makeFinal, *makePrimary)
		if err != nil {
			return nil, err
		}

		return instance, nil
	}

	stream := r.Streams.FindStream(ctx, streamID)
	if stream == nil {
		return nil, gqlerror.Errorf("Stream %s not found", streamID.String())
	}

	// check not too many instances
	if stream.InstancesCreatedCount-stream.InstancesDeletedCount >= MaxInstancesPerStream {
		return nil, gqlerror.Errorf("You cannot have more than %d instances per stream. Delete an existing instance to make room for more.", MaxInstancesPerStream)
	}

	perms := r.Permissions.StreamPermissionsForSecret(ctx, secret, stream.StreamID, stream.ProjectID, stream.Project.Public)
	if !perms.Write {
		return nil, gqlerror.Errorf("Not allowed to write to stream %s/%s", stream.Name, stream.Project.Name)
	}

	si, err := r.Streams.CreateStreamInstance(ctx, stream, version, *makePrimary)
	if err != nil {
		return nil, err
	}

	return si, nil
}

func (r *mutationResolver) UpdateStreamInstance(ctx context.Context, instanceID uuid.UUID, makeFinal *bool, makePrimary *bool) (*models.StreamInstance, error) {
	instance := r.Streams.FindStreamInstance(ctx, instanceID)
	if instance == nil {
		return nil, gqlerror.Errorf("Stream instance '%s' not found", instanceID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.StreamPermissionsForSecret(ctx, secret, instance.Stream.StreamID, instance.Stream.ProjectID, false)
	if !perms.Write {
		return nil, gqlerror.Errorf("Not allowed to write to instance %s", instanceID.String())
	}

	falseVal := false
	if makeFinal == nil {
		makeFinal = &falseVal
	}
	if makePrimary == nil {
		makePrimary = &falseVal
	}

	err := r.Streams.UpdateStreamInstance(ctx, instance.Stream, instance, *makeFinal, *makePrimary)
	if err != nil {
		return nil, gqlerror.Errorf("Error updating stream instance: %s", err.Error())
	}

	return instance, nil
}

func (r *mutationResolver) DeleteStreamInstance(ctx context.Context, instanceID uuid.UUID) (bool, error) {
	instance := r.Streams.FindStreamInstance(ctx, instanceID)
	if instance == nil {
		return false, gqlerror.Errorf("Stream instance '%s' not found", instanceID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.StreamPermissionsForSecret(ctx, secret, instance.Stream.StreamID, instance.Stream.ProjectID, false)
	if !perms.Write {
		return false, gqlerror.Errorf("Not allowed to write to instance %s", instanceID.String())
	}

	err := r.Streams.DeleteStreamInstance(ctx, instance.Stream, instance)
	if err != nil {
		return false, gqlerror.Errorf("Couldn't delete stream instance: %s", err.Error())
	}

	return true, nil
}
