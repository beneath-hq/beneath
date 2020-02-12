package resolver

import (
	"context"

	"github.com/beneath-core/pkg/middleware"

	"github.com/beneath-core/control/entity"
	"github.com/beneath-core/control/gql"
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

	secret := middleware.GetSecret(ctx)
	perms := secret.ProjectPermissions(ctx, project.ProjectID, project.Public)
	if !perms.View {
		return nil, gqlerror.Errorf("Not allowed to read project %s", name)
	}

	return project, nil
}

func (r *queryResolver) ProjectByID(ctx context.Context, projectID uuid.UUID) (*entity.Project, error) {
	project := entity.FindProject(ctx, projectID)
	if project == nil {
		return nil, gqlerror.Errorf("Project %s not found", projectID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := secret.ProjectPermissions(ctx, projectID, project.Public)
	if !perms.View {
		return nil, gqlerror.Errorf("Not allowed to read project %s", projectID.String())
	}

	return project, nil
}

func (r *mutationResolver) CreateProject(ctx context.Context, name string, displayName *string, organizationID uuid.UUID, public bool, site *string, description *string, photoURL *string) (*entity.Project, error) {
	secret := middleware.GetSecret(ctx)
	if !secret.IsUser() {
		return nil, gqlerror.Errorf("Only users can create projects")
	}

	perms := secret.OrganizationPermissions(ctx, organizationID)
	if !perms.Admin {
		return nil, gqlerror.Errorf("Not allowed to perform admin functions on organization %s", organizationID.String())
	}

	bi := entity.FindBillingInfo(ctx, organizationID)
	if bi == nil {
		return nil, gqlerror.Errorf("Could not find billing info for organization %s", organizationID.String())
	}

	if !public && !bi.BillingPlan.PrivateProjects {
		return nil, gqlerror.Errorf("Your organization's billing plan does not permit private projects")
	}

	project := &entity.Project{
		Name:           name,
		DisplayName:    DereferenceString(displayName),
		OrganizationID: organizationID,
		Site:           DereferenceString(site),
		Description:    DereferenceString(description),
		PhotoURL:       DereferenceString(photoURL),
		Public:         public,
	}

	err := project.CreateWithUser(ctx, secret.GetOwnerID(), true, true, true)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (r *mutationResolver) UpdateProject(ctx context.Context, projectID uuid.UUID, displayName *string, public *bool, site *string, description *string, photoURL *string) (*entity.Project, error) {
	project := entity.FindProject(ctx, projectID)
	if project == nil {
		return nil, gqlerror.Errorf("Project %s not found", projectID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := secret.ProjectPermissions(ctx, projectID, false)
	if !perms.Admin {
		return nil, gqlerror.Errorf("Not allowed to perform admin functions on project %s", projectID.String())
	}

	bi := entity.FindBillingInfo(ctx, project.OrganizationID)
	if bi == nil {
		return nil, gqlerror.Errorf("Could not find billing info for organization %s", project.OrganizationID.String())
	}

	if public != nil && !*public && !bi.BillingPlan.PrivateProjects {
		return nil, gqlerror.Errorf("Your organization's billing plan does not permit private projects")
	}

	err := project.UpdateDetails(ctx, displayName, public, site, description, photoURL)
	if err != nil {
		return nil, gqlerror.Errorf(err.Error())
	}

	return project, nil
}

func (r *mutationResolver) UpdateProjectOrganization(ctx context.Context, projectID uuid.UUID, organizationID uuid.UUID) (*entity.Project, error) {
	project := entity.FindProject(ctx, projectID)
	if project == nil {
		return nil, gqlerror.Errorf("Project %s not found", projectID.String())
	}

	organization := entity.FindOrganization(ctx, organizationID)
	if organization == nil {
		return nil, gqlerror.Errorf("Organization %s not found", organizationID.String())
	}

	if project.OrganizationID == organizationID {
		return nil, gqlerror.Errorf("The %s project is already owned by the %s organization", project.Name, organization.Name)
	}

	secret := middleware.GetSecret(ctx)
	projectPerms := secret.ProjectPermissions(ctx, projectID, false)
	if !projectPerms.Admin {
		return nil, gqlerror.Errorf("Not allowed to perform admin functions on project %s", projectID.String())
	}

	organizationPerms := secret.OrganizationPermissions(ctx, organizationID)
	if !organizationPerms.View {
		return nil, gqlerror.Errorf("You are not authorized for organization %s", organizationID.String())
	}

	project, err := project.UpdateOrganization(ctx, organizationID)
	if err != nil {
		return nil, gqlerror.Errorf(err.Error())
	}

	return project, nil
}

func (r *mutationResolver) AddUserToProject(ctx context.Context, username string, projectID uuid.UUID, view bool, create bool, admin bool) (*entity.User, error) {
	secret := middleware.GetSecret(ctx)
	perms := secret.ProjectPermissions(ctx, projectID, false)
	if !perms.Admin {
		return nil, gqlerror.Errorf("Not allowed to perform admin functions on project %s", projectID.String())
	}

	user := entity.FindUserByUsername(ctx, username)
	if user == nil {
		return nil, gqlerror.Errorf("No user found with that username")
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
	perms := secret.ProjectPermissions(ctx, projectID, false)
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
	perms := secret.ProjectPermissions(ctx, projectID, false)
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
