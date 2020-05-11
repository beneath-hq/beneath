package resolver

import (
	"context"

	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/gqlerror"

	"gitlab.com/beneath-hq/beneath/control/entity"
	"gitlab.com/beneath-hq/beneath/control/gql"
	"gitlab.com/beneath-hq/beneath/internal/middleware"
)

const (
	// MaxInstancesPerStream sets a limit for the number of instances for a stream at any given time
	MaxInstancesPerStream = 25
)

// Stream returns the gql.StreamResolver
func (r *Resolver) Stream() gql.StreamResolver {
	return &streamResolver{r}
}

type streamResolver struct{ *Resolver }

func (r *streamResolver) StreamID(ctx context.Context, obj *entity.Stream) (string, error) {
	return obj.StreamID.String(), nil
}

func (r *queryResolver) StreamByID(ctx context.Context, streamID uuid.UUID) (*entity.Stream, error) {
	stream := entity.FindStream(ctx, streamID)
	if stream == nil {
		return nil, gqlerror.Errorf("Stream with ID %s not found", streamID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := secret.StreamPermissions(ctx, stream.StreamID, stream.ProjectID, stream.Project.Public)
	if !perms.Read {
		return nil, gqlerror.Errorf("Not allowed to read stream with ID %s", streamID.String())
	}

	return stream, nil
}

func (r *queryResolver) StreamByOrganizationProjectAndName(ctx context.Context, organizationName string, projectName string, streamName string) (*entity.Stream, error) {
	stream := entity.FindStreamByOrganizationProjectAndName(ctx, organizationName, projectName, streamName)
	if stream == nil {
		return nil, gqlerror.Errorf("Stream %s/%s/%s not found", organizationName, projectName, streamName)
	}

	secret := middleware.GetSecret(ctx)
	perms := secret.StreamPermissions(ctx, stream.StreamID, stream.ProjectID, stream.Project.Public)
	if !perms.Read {
		return nil, gqlerror.Errorf("Not allowed to read stream %s/%s/%s", organizationName, projectName, streamName)
	}

	return stream, nil
}

func (r *queryResolver) StreamInstancesForStream(ctx context.Context, streamID uuid.UUID) ([]*entity.StreamInstance, error) {
	stream := entity.FindStream(ctx, streamID)
	if stream == nil {
		return nil, gqlerror.Errorf("Stream with ID %s not found", streamID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := secret.StreamPermissions(ctx, stream.StreamID, stream.ProjectID, stream.Project.Public)
	if !perms.Read {
		return nil, gqlerror.Errorf("Not allowed to read stream with ID %s", streamID.String())
	}

	instances := entity.FindStreamInstances(ctx, streamID)
	return instances, nil
}

func (r *mutationResolver) StageStream(ctx context.Context, organizationName string, projectName string, streamName string, schemaKind entity.StreamSchemaKind, schema string, retentionSeconds *int, enableManualWrites *bool, createPrimaryStreamInstance *bool) (*entity.Stream, error) {
	var project *entity.Project
	var stream *entity.Stream

	stream = entity.FindStreamByOrganizationProjectAndName(ctx, organizationName, projectName, streamName)
	if stream == nil {
		project = entity.FindProjectByOrganizationAndName(ctx, organizationName, projectName)
		if project == nil {
			return nil, gqlerror.Errorf("Project %s/%s not found", organizationName, projectName)
		}
		stream = &entity.Stream{
			Name:      streamName,
			ProjectID: project.ProjectID,
		}
	} else {
		project = stream.Project
	}

	secret := middleware.GetSecret(ctx)
	perms := secret.ProjectPermissions(ctx, project.ProjectID, project.Public)
	if !perms.Create {
		return nil, gqlerror.Errorf("Not allowed to create or modify resources in project %s/%s", organizationName, projectName)
	}

	err := stream.Stage(ctx, schemaKind, schema, retentionSeconds, enableManualWrites, createPrimaryStreamInstance)
	if err != nil {
		return nil, gqlerror.Errorf("Error staging stream: %s", err.Error())
	}

	return stream, nil
}

func (r *mutationResolver) DeleteStream(ctx context.Context, streamID uuid.UUID) (bool, error) {
	stream := entity.FindStream(ctx, streamID)
	if stream == nil {
		return false, gqlerror.Errorf("Stream %s not found", streamID.String())
	}

	if stream.SourceModelID != nil {
		return false, gqlerror.Errorf("Stream '%s' cannot be deleted directly because it is derived from a model (hint: delete the model)", streamID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := secret.ProjectPermissions(ctx, stream.ProjectID, false)
	if !perms.Create {
		return false, gqlerror.Errorf("Not allowed to perform admin functions in project %s", stream.Project.Name)
	}

	err := stream.Delete(ctx)
	if err != nil {
		return false, gqlerror.Errorf("%s", err.Error())
	}

	return true, nil
}

func (r *mutationResolver) CreateStreamInstance(ctx context.Context, streamID uuid.UUID) (*entity.StreamInstance, error) {
	stream := entity.FindStream(ctx, streamID)
	if stream == nil {
		return nil, gqlerror.Errorf("Stream %s not found", streamID.String())
	}

	if stream.SourceModelID != nil {
		return nil, gqlerror.Errorf("Stream '%s' cannot be manipulated directly because it is derived from a model (hint: delete the model)", stream.StreamID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := secret.StreamPermissions(ctx, stream.StreamID, stream.ProjectID, stream.Project.Public)
	if !perms.Write {
		return nil, gqlerror.Errorf("Not allowed to write to stream %s/%s", stream.Name, stream.Project.Name)
	}

	if stream.InstancesCreatedCount-stream.InstancesDeletedCount >= MaxInstancesPerStream {
		return nil, gqlerror.Errorf("You cannot have more than %d instances per stream. Delete an existing instance to make room for more.", MaxInstancesPerStream)
	}

	si, err := stream.CreateStreamInstance(ctx)
	if err != nil {
		return nil, err
	}

	return si, nil
}

func (r *mutationResolver) UpdateStreamInstance(ctx context.Context, instanceID uuid.UUID, makeFinal *bool, makePrimary *bool, deletePreviousPrimary *bool) (*entity.StreamInstance, error) {
	instance := entity.FindStreamInstance(ctx, instanceID)
	if instance == nil {
		return nil, gqlerror.Errorf("Stream instance '%s' not found", instanceID.String())
	}

	if instance.Stream.SourceModelID != nil {
		return nil, gqlerror.Errorf("Stream '%s' cannot be updated directly because it is derived from a model (hint: delete the model)", instance.StreamID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := secret.StreamPermissions(ctx, instance.Stream.StreamID, instance.Stream.ProjectID, false)
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
	if deletePreviousPrimary == nil {
		deletePreviousPrimary = &falseVal
	}

	err := instance.Stream.UpdateStreamInstance(ctx, instance, *makeFinal, *makePrimary, *deletePreviousPrimary)
	if err != nil {
		return nil, gqlerror.Errorf("Error updating stream instance: %s", err.Error())
	}

	return instance, nil
}

func (r *mutationResolver) DeleteStreamInstance(ctx context.Context, instanceID uuid.UUID) (bool, error) {
	instance := entity.FindStreamInstance(ctx, instanceID)
	if instance == nil {
		return false, gqlerror.Errorf("Stream instance '%s' not found", instanceID.String())
	}

	if instance.Stream.SourceModelID != nil {
		return false, gqlerror.Errorf("Stream '%s' cannot be deleted directly because it is derived from a model (hint: delete the model)", instance.StreamID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := secret.StreamPermissions(ctx, instance.Stream.StreamID, instance.Stream.ProjectID, false)
	if !perms.Write {
		return false, gqlerror.Errorf("Not allowed to write to instance %s", instanceID.String())
	}

	err := instance.Stream.DeleteStreamInstance(ctx, instance)
	if err != nil {
		return false, gqlerror.Errorf("Couldn't delete stream instance: %s", err.Error())
	}

	return true, nil
}
