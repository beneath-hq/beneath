package resolver

import (
	"context"

	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"gitlab.com/beneath-hq/beneath/models"
	"gitlab.com/beneath-hq/beneath/server/control/gql"
	"gitlab.com/beneath-hq/beneath/services/middleware"
)

// Project returns the gql.ProjectResolver
func (r *Resolver) Project() gql.ProjectResolver {
	return &projectResolver{r}
}

type projectResolver struct{ *Resolver }

func (r *projectResolver) ProjectID(ctx context.Context, obj *models.Project) (string, error) {
	return obj.ProjectID.String(), nil
}

func (r *queryResolver) ExploreProjects(ctx context.Context) ([]*models.Project, error) {
	return r.Projects.ExploreProjects(ctx), nil
}

func (r *queryResolver) ProjectsForUser(ctx context.Context, userID uuid.UUID) ([]*models.Project, error) {
	secret := middleware.GetSecret(ctx)
	if !(secret.IsUser() && secret.GetOwnerID() == userID) {
		return nil, gqlerror.Errorf("ProjectsForUser can only be called for the calling user")
	}
	return r.Projects.FindProjectsForUser(ctx, userID), nil
}

func (r *queryResolver) ProjectByOrganizationAndName(ctx context.Context, organizationName string, projectName string) (*models.Project, error) {
	project := r.Projects.FindProjectByOrganizationAndName(ctx, organizationName, projectName)
	if project == nil {
		return nil, gqlerror.Errorf("Project %s/%s not found", organizationName, projectName)
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.ProjectPermissionsForSecret(ctx, secret, project.ProjectID, project.Public)
	if !perms.View {
		return nil, gqlerror.Errorf("Not allowed to read project %s/%s", organizationName, projectName)
	}

	return projectWithPermissions(project, perms), nil
}

func (r *queryResolver) ProjectByID(ctx context.Context, projectID uuid.UUID) (*models.Project, error) {
	project := r.Projects.FindProject(ctx, projectID)
	if project == nil {
		return nil, gqlerror.Errorf("Project %s not found", projectID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.ProjectPermissionsForSecret(ctx, secret, projectID, project.Public)
	if !perms.View {
		return nil, gqlerror.Errorf("Not allowed to read project %s", projectID.String())
	}

	return projectWithPermissions(project, perms), nil
}

func (r *queryResolver) ProjectMembers(ctx context.Context, projectID uuid.UUID) ([]*models.ProjectMember, error) {
	project := r.Projects.FindProject(ctx, projectID)
	if project == nil {
		return nil, gqlerror.Errorf("Project %s not found", projectID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.ProjectPermissionsForSecret(ctx, secret, projectID, project.Public)
	if !perms.View {
		return nil, gqlerror.Errorf("You're not allowed to see the members of project %s", projectID.String())
	}

	return r.Projects.FindProjectMembers(ctx, projectID)
}

func (r *mutationResolver) CreateProject(ctx context.Context, input gql.CreateProjectInput) (*models.Project, error) {
	secret := middleware.GetSecret(ctx)
	orgPerms := r.Permissions.OrganizationPermissionsForSecret(ctx, secret, input.OrganizationID)
	if !orgPerms.Create {
		return nil, gqlerror.Errorf("Not allowed to perform admin functions on organization %s", input.OrganizationID.String())
	}

	organization := r.Organizations.FindOrganization(ctx, input.OrganizationID)
	if organization == nil {
		return nil, gqlerror.Errorf("Organization %s not found", input.OrganizationID.String())
	}

	project := &models.Project{
		Name:           input.ProjectName,
		OrganizationID: organization.OrganizationID,
		Organization:   organization,
	}

	perms := models.ProjectPermissions{
		View:   true,
		Create: true,
		Admin:  true,
	}

	err := r.Projects.CreateWithUser(ctx, project, input.DisplayName, input.Public, input.Description, input.Site, input.PhotoURL, secret.GetOwnerID(), perms)
	if err != nil {
		return nil, gqlerror.Errorf("Error creating project: %s", err.Error())
	}

	user := r.Users.FindUser(ctx, secret.GetOwnerID())
	if user != nil {
		project.Users = append(project.Users, user)
	}

	return projectWithPermissions(project, perms), nil
}

func (r *mutationResolver) UpdateProject(ctx context.Context, input gql.UpdateProjectInput) (*models.Project, error) {
	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.ProjectPermissionsForSecret(ctx, secret, input.ProjectID, false)
	if !perms.Admin {
		return nil, gqlerror.Errorf("Not allowed to perform admin functions on project %s", input.ProjectID.String())
	}

	project := r.Projects.FindProject(ctx, input.ProjectID)
	if project == nil {
		return nil, gqlerror.Errorf("Project %s not found", input.ProjectID.String())
	}

	err := r.Projects.Update(ctx, project, input.DisplayName, input.Public, input.Description, input.Site, input.PhotoURL)
	if err != nil {
		return nil, gqlerror.Errorf("Error updating project: %s", err.Error())
	}

	return projectWithPermissions(project, perms), nil
}

func (r *mutationResolver) DeleteProject(ctx context.Context, input gql.DeleteProjectInput) (bool, error) {
	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.ProjectPermissionsForSecret(ctx, secret, input.ProjectID, false)
	if !perms.Admin {
		return false, gqlerror.Errorf("Not allowed to perform admin functions in project %s", input.ProjectID.String())
	}

	project := r.Projects.FindProject(ctx, input.ProjectID)
	if project == nil {
		return false, gqlerror.Errorf("Project %s not found", input.ProjectID.String())
	}

	err := r.Projects.Delete(ctx, project)
	if err != nil {
		return false, gqlerror.Errorf(err.Error())
	}

	return true, nil
}

func projectWithPermissions(p *models.Project, perms models.ProjectPermissions) *models.Project {
	if perms.View || perms.Create || perms.Admin {
		p.Permissions = &models.PermissionsUsersProjects{
			View:   perms.View,
			Create: perms.Create,
			Admin:  perms.Admin,
		}
	}
	return p
}
