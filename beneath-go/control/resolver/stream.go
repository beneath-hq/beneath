package resolver

import (
	"context"

	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/gqlerror"

	"github.com/beneath-core/beneath-go/control/gql"
	"github.com/beneath-core/beneath-go/control/model"
	"github.com/beneath-core/beneath-go/core/middleware"
)

// Stream returns the gql.StreamResolver
func (r *Resolver) Stream() gql.StreamResolver {
	return &streamResolver{r}
}

type streamResolver struct{ *Resolver }

func (r *streamResolver) StreamID(ctx context.Context, obj *model.Stream) (string, error) {
	return obj.StreamID.String(), nil
}

func (r *queryResolver) Stream(ctx context.Context, name string, projectName string) (*model.Stream, error) {
	stream := model.FindStreamByNameAndProject(ctx, name, projectName)
	if stream == nil {
		return nil, gqlerror.Errorf("Stream %s/%s not found", projectName, name)
	}

	secret := middleware.GetSecret(ctx)
	if !secret.ReadsProject(stream.ProjectID) {
		return nil, gqlerror.Errorf("Not allowed to read stream %s/%s", projectName, name)
	}

	return stream, nil
}

func (r *mutationResolver) CreateExternalStream(ctx context.Context, projectID uuid.UUID, schema string, batch bool, manual bool) (*model.Stream, error) {
	secret := middleware.GetSecret(ctx)
	if !secret.EditsProject(projectID) {
		return nil, gqlerror.Errorf("Not allowed to edit project %s", projectID)
	}

	stream := &model.Stream{
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
	return model.FindStream(ctx, stream.StreamID), nil
}

func (r *mutationResolver) UpdateStream(ctx context.Context, streamID uuid.UUID, schema *string, manual *bool) (*model.Stream, error) {
	stream := model.FindStream(ctx, streamID)
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
