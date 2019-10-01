package resolver

import (
	"context"

	"github.com/beneath-core/beneath-go/core/middleware"

	"github.com/beneath-core/beneath-go/control/entity"
	"github.com/beneath-core/beneath-go/control/gql"
	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/gqlerror"
)

// Project returns the gql.ProjectResolver
func (r *Resolver) Project() gql.ProjectResolver {
	return &projectResolver{r}
}

type projectResolver struct{ *Resolver }

func (r *projectResolver) ProjectID(ctx context.Context, obj *entity.Project) (string, error) {
	return obj.ProjectID.String(), nil
}

func (r *queryResolver) ExploreProjects(ctx context.Context) ([]*entity.Project, error) {
	return entity.FindProjects(ctx), nil
}

func (r *queryResolver) ProjectByName(ctx context.Context, name string) (*entity.Project, error) {
	project := entity.FindProjectByName(ctx, name)
	if project == nil {
		return nil, gqlerror.Errorf("Project %s not found", name)
	}

	if !project.Public {
		secret := middleware.GetSecret(ctx)
		perms := secret.ProjectPermissions(ctx, project.ProjectID)
		if !perms.View {
			return nil, gqlerror.Errorf("Not allowed to read project %s", name)
		}
	}

	return project, nil
}

func (r *queryResolver) ProjectByID(ctx context.Context, projectID uuid.UUID) (*entity.Project, error) {
	project := entity.FindProject(ctx, projectID)
	if project == nil {
		return nil, gqlerror.Errorf("Project %s not found", projectID.String())
	}

	if !project.Public {
		secret := middleware.GetSecret(ctx)
		perms := secret.ProjectPermissions(ctx, projectID)
		if !perms.View {
			return nil, gqlerror.Errorf("Not allowed to read project %s", projectID.String())
		}
	}

	return project, nil
}

func (r *mutationResolver) CreateProject(ctx context.Context, name string, displayName string, organizationID uuid.UUID, site *string, description *string, photoURL *string) (*entity.Project, error) {
	secret := middleware.GetSecret(ctx)
	if !secret.IsUser() {
		return nil, gqlerror.Errorf("Not allowed to create project")
	}

	project := &entity.Project{
		Name:           name,
		DisplayName:    displayName,
		OrganizationID: organizationID,
		Site:           DereferenceString(site),
		Description:    DereferenceString(description),
		PhotoURL:       DereferenceString(photoURL),
		Public:         true,
	}

	err := project.CreateWithUser(ctx, *secret.UserID, true, true, true)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (r *mutationResolver) UpdateProject(ctx context.Context, projectID uuid.UUID, displayName *string, site *string, description *string, photoURL *string) (*entity.Project, error) {
	secret := middleware.GetSecret(ctx)
	perms := secret.ProjectPermissions(ctx, projectID)
	if !perms.Admin {
		return nil, gqlerror.Errorf("Not allowed to perform admin functions on project %s", projectID.String())
	}

	project := entity.FindProject(ctx, projectID)
	if project == nil {
		return nil, gqlerror.Errorf("Project %s not found", projectID.String())
	}

	err := project.UpdateDetails(ctx, displayName, site, description, photoURL)
	if err != nil {
		return nil, gqlerror.Errorf(err.Error())
	}

	return project, nil
}

func (r *mutationResolver) AddUserToProject(ctx context.Context, email string, projectID uuid.UUID, view bool, create bool, admin bool) (*entity.User, error) {
	secret := middleware.GetSecret(ctx)
	perms := secret.ProjectPermissions(ctx, projectID)
	if !perms.Admin {
		return nil, gqlerror.Errorf("Not allowed to perform admin functions on project %s", projectID.String())
	}

	user := entity.FindUserByEmail(ctx, email)
	if user == nil {
		return nil, gqlerror.Errorf("No user found with that email")
	}

	project := &entity.Project{
		ProjectID: projectID,
	}

	err := project.AddUser(ctx, user.UserID, view, create, admin)
	if err != nil {
		return nil, gqlerror.Errorf(err.Error())
	}

	return user, nil
}

func (r *mutationResolver) RemoveUserFromProject(ctx context.Context, userID uuid.UUID, projectID uuid.UUID) (bool, error) {
	secret := middleware.GetSecret(ctx)
	perms := secret.ProjectPermissions(ctx, projectID)
	if !perms.Admin {
		return false, gqlerror.Errorf("Not allowed to perform admin functions in project %s", projectID.String())
	}

	project := entity.FindProject(ctx, projectID)
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

func (r *mutationResolver) DeleteProject(ctx context.Context, projectID uuid.UUID) (bool, error) {
	secret := middleware.GetSecret(ctx)
	perms := secret.ProjectPermissions(ctx, projectID)
	if !perms.Admin {
		return false, gqlerror.Errorf("Not allowed to perform admin functions in project %s", projectID.String())
	}

	project := entity.FindProject(ctx, projectID)
	if project == nil {
		return false, gqlerror.Errorf("Project %s not found", projectID.String())
	}

	err := project.Delete(ctx)
	if err != nil {
		return false, gqlerror.Errorf(err.Error())
	}

	return true, nil
}
