package resolver

import (
	"context"

	"github.com/beneath-core/beneath-go/core/middleware"

	"github.com/beneath-core/beneath-go/control/gql"
	"github.com/beneath-core/beneath-go/control/model"
	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/gqlerror"
)

// Project returns the gql.ProjectResolver
func (r *Resolver) Project() gql.ProjectResolver {
	return &projectResolver{r}
}

type projectResolver struct{ *Resolver }

func (r *projectResolver) ProjectID(ctx context.Context, obj *model.Project) (string, error) {
	return obj.ProjectID.String(), nil
}

func (r *queryResolver) ExploreProjects(ctx context.Context) ([]*model.Project, error) {
	return model.FindProjects(ctx), nil
}

func (r *queryResolver) ProjectByName(ctx context.Context, name string) (*model.Project, error) {
	project := model.FindProjectByName(ctx, name)
	if project == nil {
		return nil, gqlerror.Errorf("Project %s not found", name)
	}

	secret := middleware.GetSecret(ctx)
	if !secret.ReadsProject(project.ProjectID) {
		return nil, gqlerror.Errorf("Not allowed to read project %s", name)
	}

	return project, nil
}

func (r *queryResolver) ProjectByID(ctx context.Context, projectID uuid.UUID) (*model.Project, error) {
	secret := middleware.GetSecret(ctx)
	if !secret.ReadsProject(projectID) {
		return nil, gqlerror.Errorf("Not allowed to read project %s", projectID.String())
	}

	project := model.FindProject(ctx, projectID)
	if project == nil {
		return nil, gqlerror.Errorf("Project %s not found", projectID.String())
	}

	return project, nil
}

func (r *mutationResolver) CreateProject(ctx context.Context, name string, displayName string, site *string, description *string, photoURL *string) (*model.Project, error) {
	secret := middleware.GetSecret(ctx)
	if !secret.IsPersonal() {
		return nil, gqlerror.Errorf("Not allowed to create project")
	}

	project := &model.Project{
		Name:        name,
		DisplayName: displayName,
		Site:        DereferenceString(site),
		Description: DereferenceString(description),
		PhotoURL:    DereferenceString(photoURL),
		Public:      true,
	}

	err := project.CreateWithUser(ctx, *secret.UserID)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (r *mutationResolver) UpdateProject(ctx context.Context, projectID uuid.UUID, displayName *string, site *string, description *string, photoURL *string) (*model.Project, error) {
	secret := middleware.GetSecret(ctx)
	if !secret.EditsProject(projectID) {
		return nil, gqlerror.Errorf("Not allowed to edit project %s", projectID.String())
	}

	project := model.FindProject(ctx, projectID)
	if project == nil {
		return nil, gqlerror.Errorf("Project %s not found", projectID.String())
	}

	err := project.UpdateDetails(ctx, displayName, site, description, photoURL)
	if err != nil {
		return nil, gqlerror.Errorf(err.Error())
	}

	return project, nil
}

func (r *mutationResolver) AddUserToProject(ctx context.Context, email string, projectID uuid.UUID) (*model.User, error) {
	secret := middleware.GetSecret(ctx)
	if !secret.EditsProject(projectID) {
		return nil, gqlerror.Errorf("Not allowed to edit project %s", projectID.String())
	}

	user := model.FindUserByEmail(ctx, email)
	if user == nil {
		return nil, gqlerror.Errorf("No user found with that email")
	}

	project := &model.Project{
		ProjectID: projectID,
	}

	err := project.AddUser(ctx, user.UserID)
	if err != nil {
		return nil, gqlerror.Errorf(err.Error())
	}

	return user, nil
}

func (r *mutationResolver) RemoveUserFromProject(ctx context.Context, userID uuid.UUID, projectID uuid.UUID) (bool, error) {
	secret := middleware.GetSecret(ctx)
	if !secret.EditsProject(projectID) {
		return false, gqlerror.Errorf("Not allowed to edit project %s", projectID.String())
	}

	project := model.FindProject(ctx, projectID)
	if project == nil {
		return false, gqlerror.Errorf("Project %s not found", projectID.String())
	}

	if len(project.Users) < 2 {
		return false, gqlerror.Errorf("Can't remove last member of project")
	}

	err := project.RemoveUser(ctx, userID)
	if err != nil {
		return false, gqlerror.Errorf(err.Error())
	}

	return true, nil
}
