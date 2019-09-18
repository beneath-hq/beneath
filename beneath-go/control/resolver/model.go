package resolver

import (
	"context"

	"github.com/beneath-core/beneath-go/control/entity"
	"github.com/beneath-core/beneath-go/control/gql"
	"github.com/beneath-core/beneath-go/core/middleware"
	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/gqlerror"
)

// Model returns the gql.ModelResolver
func (r *Resolver) Model() gql.ModelResolver {
	return &modelResolver{r}
}

type modelResolver struct{ *Resolver }

func (r *modelResolver) ModelID(ctx context.Context, obj *entity.Model) (string, error) {
	return obj.ModelID.String(), nil
}

func (r *modelResolver) Kind(ctx context.Context, obj *entity.Model) (string, error) {
	return string(obj.Kind), nil
}

func (r *queryResolver) Model(ctx context.Context, name string, projectName string) (*entity.Model, error) {
	model := entity.FindModelByNameAndProject(ctx, name, projectName)
	if model == nil {
		return nil, gqlerror.Errorf("Model %s/%s not found", projectName, name)
	}

	secret := middleware.GetSecret(ctx)
	if !secret.ReadsProject(model.ProjectID) {
		return nil, gqlerror.Errorf("Not allowed to read model %s/%s", projectName, name)
	}

	return model, nil
}

func (r *mutationResolver) CreateModel(ctx context.Context, input gql.CreateModelInput) (*entity.Model, error) {
	secret := middleware.GetSecret(ctx)
	if !secret.EditsProject(input.ProjectID) {
		return nil, gqlerror.Errorf("Not allowed to edit project %s", input.ProjectID)
	}

	kind, ok := entity.ParseModelKind(input.Kind)
	if !ok {
		return nil, gqlerror.Errorf("Unrecognized model kind '%s'", input.Kind)
	}

	model := &entity.Model{
		Name:        input.Name,
		Description: DereferenceString(input.Description),
		SourceURL:   DereferenceString(input.SourceURL),
		Kind:        kind,
		ProjectID:   input.ProjectID,
	}

	err := model.CompileAndCreate(ctx, input.InputStreamIDs, input.OutputStreamSchemas)
	if err != nil {
		return nil, gqlerror.Errorf(err.Error())
	}

	// done (using FindModel to get relations correctly)
	return entity.FindModel(ctx, model.ModelID), nil
}

func (r *mutationResolver) UpdateModel(ctx context.Context, input gql.UpdateModelInput) (*entity.Model, error) {
	// get model
	model := entity.FindModel(ctx, input.ModelID)
	if model == nil {
		return nil, gqlerror.Errorf("Model %s not found", input.ModelID.String())
	}

	// check allowed to edit
	secret := middleware.GetSecret(ctx)
	if !secret.EditsProject(model.ProjectID) {
		return nil, gqlerror.Errorf("Not allowed to edit project '%s'", model.ProjectID)
	}

	// update model
	if input.SourceURL != nil {
		model.SourceURL = DereferenceString(input.SourceURL)
	}
	if input.Description != nil {
		model.Description = DereferenceString(input.Description)
	}

	// compile and update
	err := model.CompileAndUpdate(ctx, input.InputStreamIDs, input.OutputStreamSchemas)
	if err != nil {
		return nil, gqlerror.Errorf(err.Error())
	}

	// done (using FindModel to get relations correctly)
	return entity.FindModel(ctx, model.ModelID), nil
}

func (r *mutationResolver) DeleteModel(ctx context.Context, modelID uuid.UUID) (bool, error) {
	// get model
	model := entity.FindModel(ctx, modelID)
	if model == nil {
		return false, gqlerror.Errorf("Model %s not found", modelID.String())
	}

	// check allowed to edit
	secret := middleware.GetSecret(ctx)
	if !secret.EditsProject(model.ProjectID) {
		return false, gqlerror.Errorf("Not allowed to edit project '%s'", model.ProjectID)
	}

	// delete model
	err := model.Delete(ctx)
	if err != nil {
		return false, gqlerror.Errorf(err.Error())
	}

	return true, nil
}

func (r *mutationResolver) NewBatch(ctx context.Context, modelID uuid.UUID) ([]*entity.StreamInstance, error) {
	// what about new batches of root streams?
	panic("not implemented")
}

func (r *mutationResolver) CommitBatch(ctx context.Context, instanceIDs []*uuid.UUID) (bool, error) {
	panic("not implemented")
}
