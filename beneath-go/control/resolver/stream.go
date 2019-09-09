package resolver

import (
	"context"

	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/gqlerror"

	"github.com/beneath-core/beneath-go/control/gql"
	"github.com/beneath-core/beneath-go/control/entity"
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
	if !secret.ReadsProject(stream.ProjectID) {
		return nil, gqlerror.Errorf("Not allowed to read stream %s/%s", projectName, name)
	}

	return stream, nil
}

func (r *mutationResolver) CreateExternalStream(ctx context.Context, projectID uuid.UUID, schema string, batch bool, manual bool) (*entity.Stream, error) {
	secret := middleware.GetSecret(ctx)
	if !secret.EditsProject(projectID) {
		return nil, gqlerror.Errorf("Not allowed to edit project %s", projectID)
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

func (r *mutationResolver) UpdateStream(ctx context.Context, streamID uuid.UUID, schema *string, manual *bool) (*entity.Stream, error) {
	stream := entity.FindStream(ctx, streamID)
	if stream == nil {
		return nil, gqlerror.Errorf("Stream %s not found", streamID.String())
	}

	secret := middleware.GetSecret(ctx)
	if !secret.EditsProject(stream.ProjectID) {
		return nil, gqlerror.Errorf("Not allowed to update stream in project %s", stream.Project.Name)
	}

	err := stream.UpdateDetails(ctx, schema, manual)
	if err != nil {
		return nil, gqlerror.Errorf(err.Error())
	}

	return stream, nil
}
