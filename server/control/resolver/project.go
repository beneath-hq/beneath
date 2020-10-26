package resolver

import (
	"context"
	"fmt"

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

func (r *mutationResolver) StageProject(ctx context.Context, organizationName string, projectName string, displayName *string, public *bool, description *string, site *string, photoURL *string) (*models.Project, error) {
	var organization *models.Organization
	var project *models.Project

	// find or prepare for creating project
	project = r.Projects.FindProjectByOrganizationAndName(ctx, organizationName, projectName)
	if project == nil {
		organization = r.Organizations.FindOrganizationByName(ctx, organizationName)
		if organization == nil {
			return nil, gqlerror.Errorf("Organization %s not found", organizationName)
		}
		project = &models.Project{
			Name:           projectName,
			OrganizationID: organization.OrganizationID,
			Organization:   organization,
		}
	} else {
		organization = project.Organization
	}

	// determine if we're creating project
	create := project.ProjectID == uuid.Nil

	// check permissions (on org if we're creating the project)
	var perms models.ProjectPermissions
	secret := middleware.GetSecret(ctx)
	if create {
		orgPerms := r.Permissions.OrganizationPermissionsForSecret(ctx, secret, organization.OrganizationID)
		if !orgPerms.Create {
			return nil, gqlerror.Errorf("Not allowed to perform admin functions on organization %s", organizationName)
		}
		perms = models.ProjectPermissions{
			View:   true,
			Create: true,
			Admin:  true,
		}
	} else {
		perms = r.Permissions.ProjectPermissionsForSecret(ctx, secret, project.ProjectID, false)
		if !perms.Admin {
			return nil, gqlerror.Errorf("Not allowed to perform admin functions on project %s/%s", organizationName, projectName)
		}
	}

	// default public to false upon creation
	if create && public == nil {
		falseVal := false
		public = &falseVal
	}

	err := r.Projects.StageWithUser(ctx, project, displayName, public, description, site, photoURL, secret.GetOwnerID(), perms)
	if err != nil {
		return nil, gqlerror.Errorf("Error staging project: %s", err.Error())
	}

	// on create, refetching project to include user
	if create {
		project = r.Projects.FindProject(ctx, project.ProjectID)
		if project == nil {
			panic(fmt.Errorf("expected project with ID %s to exist", project.ProjectID.String()))
		}
	}

	return projectWithPermissions(project, perms), nil
}

func (r *mutationResolver) DeleteProject(ctx context.Context, projectID uuid.UUID) (bool, error) {
	secret := middleware.GetSecret(ctx)
	perms := r.Permissions.ProjectPermissionsForSecret(ctx, secret, projectID, false)
	if !perms.Admin {
		return false, gqlerror.Errorf("Not allowed to perform admin functions in project %s", projectID.String())
	}

	project := r.Projects.FindProject(ctx, projectID)
	if project == nil {
		return false, gqlerror.Errorf("Project %s not found", projectID.String())
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
