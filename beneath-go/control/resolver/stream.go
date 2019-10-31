package resolver

import (
	"context"

	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/gqlerror"

	"github.com/beneath-core/beneath-go/control/entity"
	"github.com/beneath-core/beneath-go/control/gql"
	"github.com/beneath-core/beneath-go/core/middleware"
)

// Stream returns the gql.StreamResolver
func (r *Resolver) Stream() gql.StreamResolver {
	return &streamResolver{r}
}

type streamResolver struct{ *Resolver }

func (r *streamResolver) StreamID(ctx context.Context, obj *entity.Stream) (string, error) {
	return obj.StreamID.String(), nil
}

func (r *queryResolver) Stream(ctx context.Context, name string, projectName string) (*entity.Stream, error) {
	stream := entity.FindStreamByNameAndProject(ctx, name, projectName)
	if stream == nil {
		return nil, gqlerror.Errorf("Stream %s/%s not found", projectName, name)
	}

	secret := middleware.GetSecret(ctx)
	perms := secret.StreamPermissions(ctx, stream.StreamID, stream.ProjectID, stream.Project.Public, stream.External)
	if !perms.Read {
		return nil, gqlerror.Errorf("Not allowed to read stream %s/%s", projectName, name)
	}

	return stream, nil
}

func (r *mutationResolver) CreateExternalStream(ctx context.Context, projectID uuid.UUID, schema string, batch bool, manual bool) (*entity.Stream, error) {
	secret := middleware.GetSecret(ctx)
	perms := secret.ProjectPermissions(ctx, projectID, false)
	if !perms.Create {
		return nil, gqlerror.Errorf("Not allowed to create content in project %s", projectID)
	}

	stream := &entity.Stream{
		Schema:    schema,
		External:  true,
		Batch:     batch,
		Manual:    manual,
		ProjectID: projectID,
	}

	err := stream.CompileAndCreate(ctx)
	if err != nil {
		return nil, gqlerror.Errorf(err.Error())
	}

	// done (using FindStream to get relations correctly)
	return entity.FindStream(ctx, stream.StreamID), nil
}

func (r *mutationResolver) UpdateExternalStream(ctx context.Context, streamID uuid.UUID, schema *string, manual *bool) (*entity.Stream, error) {
	stream := entity.FindStream(ctx, streamID)
	if stream == nil {
		return nil, gqlerror.Errorf("Stream %s not found", streamID.String())
	}

	if !stream.External {
		return nil, gqlerror.Errorf("Stream '%s' is not a root stream", streamID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := secret.ProjectPermissions(ctx, stream.ProjectID, false)
	if !perms.Create {
		return nil, gqlerror.Errorf("Not allowed to update stream in project %s", stream.Project.Name)
	}

	err := stream.CompileAndUpdate(ctx, schema, manual)
	if err != nil {
		return nil, gqlerror.Errorf(err.Error())
	}

	return stream, nil
}

func (r *mutationResolver) DeleteExternalStream(ctx context.Context, streamID uuid.UUID) (bool, error) {
	stream := entity.FindStream(ctx, streamID)
	if stream == nil {
		return false, gqlerror.Errorf("Stream %s not found", streamID.String())
	}

	if !stream.External {
		return false, gqlerror.Errorf("Stream '%s' is not a root stream", streamID.String())
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

func (r *mutationResolver) CreateExternalStreamBatch(ctx context.Context, streamID uuid.UUID) (*entity.StreamInstance, error) {
	stream := entity.FindStream(ctx, streamID)
	if stream == nil {
		return nil, gqlerror.Errorf("Stream %s not found", streamID.String())
	}

	if !stream.External {
		return nil, gqlerror.Errorf("Stream '%s' is not a root stream", streamID.String())
	}

	if !stream.Batch {
		return nil, gqlerror.Errorf("Stream '%s' is not a batch stream", streamID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := secret.StreamPermissions(ctx, stream.StreamID, stream.ProjectID, stream.Project.Public, stream.External)
	if !perms.Write {
		return nil, gqlerror.Errorf("Not allowed to write to stream %s/%s", stream.Name, stream.Project.Name)
	}

	si, err := stream.CreateStreamInstance(ctx)
	if err != nil {
		return nil, err
	}

	return si, nil
}

func (r *mutationResolver) CommitExternalStreamBatch(ctx context.Context, instanceID uuid.UUID) (bool, error) {
	instance := entity.FindStreamInstance(ctx, instanceID)
	if instance == nil {
		return false, gqlerror.Errorf("Stream instance '%s' not found", instanceID.String())
	}

	if !instance.Stream.External {
		return false, gqlerror.Errorf("Stream '%s/%s' is not a root stream", instance.Stream.Name, instance.Stream.Project.Name)
	}

	if !instance.Stream.Batch {
		return false, gqlerror.Errorf("Stream '%s' is not a batch stream", instance.StreamID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := secret.StreamPermissions(ctx, instance.Stream.StreamID, instance.Stream.ProjectID, false, instance.Stream.External)
	if !perms.Write {
		return false, gqlerror.Errorf("Not allowed to write to instance %s", instanceID.String())
	}

	err := instance.Stream.CommitStreamInstance(ctx, instance)
	if err != nil {
		return false, gqlerror.Errorf("%s", err.Error())
	}

	return true, nil
}

func (r *mutationResolver) ClearPendingExternalStreamBatches(ctx context.Context, streamID uuid.UUID) (bool, error) {
	stream := entity.FindStream(ctx, streamID)
	if stream == nil {
		return false, gqlerror.Errorf("Stream %s not found", streamID.String())
	}

	if !stream.External {
		return false, gqlerror.Errorf("Stream '%s' is not a root stream", streamID.String())
	}

	if !stream.Batch {
		return false, gqlerror.Errorf("Stream '%s' is not a batch stream", streamID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := secret.StreamPermissions(ctx, stream.StreamID, stream.ProjectID, stream.Project.Public, stream.External)
	if !perms.Write {
		return false, gqlerror.Errorf("Not allowed to write stream %s/%s", stream.Name, stream.Project.Name)
	}

	err := stream.ClearPendingBatches(ctx)
	if err != nil {
		return false, gqlerror.Errorf("%s", err.Error())
	}

	return true, nil
}
