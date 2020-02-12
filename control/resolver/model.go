package resolver

import (
	"context"

	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/gqlerror"

	"github.com/beneath-core/control/entity"
	"github.com/beneath-core/control/gql"
	"github.com/beneath-core/pkg/middleware"
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
	perms := secret.ProjectPermissions(ctx, model.ProjectID, model.Project.Public)
	if !perms.View {
		return nil, gqlerror.Errorf("Not allowed to read model %s/%s", projectName, name)
	}

	return model, nil
}

func (r *mutationResolver) CreateModel(ctx context.Context, input gql.CreateModelInput) (*entity.Model, error) {
	secret := middleware.GetSecret(ctx)
	perms := secret.ProjectPermissions(ctx, input.ProjectID, false)
	if !perms.Create {
		return nil, gqlerror.Errorf("Not allowed to modify resources in project %s", input.ProjectID)
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

	err := model.CompileAndCreate(ctx, input.InputStreamIDs, input.OutputStreamSchemas, int64(input.ReadQuota), int64(input.WriteQuota))
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
	perms := secret.ProjectPermissions(ctx, model.ProjectID, false)
	if !perms.Create {
		return nil, gqlerror.Errorf("Not allowed to modify resources in project '%s'", model.ProjectID)
	}

	// update model
	if input.SourceURL != nil {
		model.SourceURL = DereferenceString(input.SourceURL)
	}
	if input.Description != nil {
		model.Description = DereferenceString(input.Description)
	}

	// compile and update
	err := model.CompileAndUpdate(ctx, input.InputStreamIDs, input.OutputStreamSchemas, int64(*input.ReadQuota), int64(*input.WriteQuota))
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
	perms := secret.ProjectPermissions(ctx, model.ProjectID, false)
	if !perms.Create {
		return false, gqlerror.Errorf("Not allowed to modify resources in project '%s'", model.ProjectID)
	}

	// delete model
	err := model.Delete(ctx)
	if err != nil {
		return false, gqlerror.Errorf(err.Error())
	}

	return true, nil
}

func (r *mutationResolver) CreateModelBatch(ctx context.Context, modelID uuid.UUID) ([]*entity.StreamInstance, error) {
	// get model
	model := entity.FindModel(ctx, modelID)
	if model == nil {
		return nil, gqlerror.Errorf("Model '%s' not found", modelID.String())
	}

	// check permissions
	secret := middleware.GetSecret(ctx)
	if !secret.ManagesModelBatches(model) {
		return nil, gqlerror.Errorf("Not allowed to modify model '%s'", model.ModelID)
	}

	// check is batch
	if model.Kind != entity.ModelKindBatch {
		return nil, gqlerror.Errorf("Cannot manage batches for streaming or microbatch model")
	}

	// make instances
	instances, err := model.CreateBatch(ctx)
	if err != nil {
		return nil, gqlerror.Errorf("Error creating model batch: %s", err.Error())
	}

	return instances, nil
}

func (r *mutationResolver) CommitModelBatch(ctx context.Context, modelID uuid.UUID, instanceIDs []uuid.UUID) (bool, error) {
	// get model
	model := entity.FindModel(ctx, modelID)
	if model == nil {
		return false, gqlerror.Errorf("Model '%s' not found", modelID.String())
	}

	// check allowed to edit
	secret := middleware.GetSecret(ctx)
	if !secret.ManagesModelBatches(model) {
		return false, gqlerror.Errorf("Not allowed to modify model '%s'", model.ModelID)
	}

	// check is batch
	if model.Kind != entity.ModelKindBatch {
		return false, gqlerror.Errorf("Cannot manage batches for streaming or microbatch model")
	}

	// commit
	err := model.CommitBatch(ctx, instanceIDs)
	if err != nil {
		return false, gqlerror.Errorf("Error committing batch: %s", err.Error())
	}

	return true, nil
}

func (r *mutationResolver) ClearPendingModelBatches(ctx context.Context, modelID uuid.UUID) (bool, error) {
	// get model
	model := entity.FindModel(ctx, modelID)
	if model == nil {
		return false, gqlerror.Errorf("Model '%s' not found", modelID.String())
	}

	// check allowed to edit
	secret := middleware.GetSecret(ctx)
	if !secret.ManagesModelBatches(model) {
		return false, gqlerror.Errorf("Not allowed to modify model '%s'", model.ModelID)
	}

	// check is batch
	if model.Kind != entity.ModelKindBatch {
		return false, gqlerror.Errorf("Cannot manage batches for streaming or microbatch model")
	}

	// delete
	err := model.ClearPendingBatches(ctx)
	if err != nil {
		return false, gqlerror.Errorf("Error committing batch: %s", err.Error())
	}

	return true, nil
}
