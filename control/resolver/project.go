package resolver

import (
	"context"
	"fmt"

	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/gqlerror"

	"gitlab.com/beneath-hq/beneath/control/entity"
	"gitlab.com/beneath-hq/beneath/control/gql"
	"gitlab.com/beneath-hq/beneath/internal/middleware"
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
	return entity.ExploreProjects(ctx), nil
}

func (r *queryResolver) ProjectsForUser(ctx context.Context, userID uuid.UUID) ([]*entity.Project, error) {
	secret := middleware.GetSecret(ctx)
	if !(secret.IsUser() && secret.GetOwnerID() == userID) {
		return nil, gqlerror.Errorf("ProjectsForUser can only be called for the calling user")
	}
	return entity.FindProjectsForUser(ctx, userID), nil
}

func (r *queryResolver) ProjectByOrganizationAndName(ctx context.Context, organizationName string, projectName string) (*entity.Project, error) {
	project := entity.FindProjectByOrganizationAndName(ctx, organizationName, projectName)
	if project == nil {
		return nil, gqlerror.Errorf("Project %s/%s not found", organizationName, projectName)
	}

	secret := middleware.GetSecret(ctx)
	perms := secret.ProjectPermissions(ctx, project.ProjectID, project.Public)
	if !perms.View {
		return nil, gqlerror.Errorf("Not allowed to read project %s/%s", organizationName, projectName)
	}

	return projectWithPermissions(project, perms), nil
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

	return projectWithPermissions(project, perms), nil
}

func (r *queryResolver) ProjectMembers(ctx context.Context, projectID uuid.UUID) ([]*entity.ProjectMember, error) {
	project := entity.FindProject(ctx, projectID)
	if project == nil {
		return nil, gqlerror.Errorf("Project %s not found", projectID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := secret.ProjectPermissions(ctx, projectID, project.Public)
	if !perms.View {
		return nil, gqlerror.Errorf("You're not allowed to see the members of project %s", projectID.String())
	}

	return entity.FindProjectMembers(ctx, projectID)
}

func (r *mutationResolver) StageProject(ctx context.Context, organizationName string, projectName string, displayName *string, public *bool, description *string, site *string, photoURL *string) (*entity.Project, error) {
	var organization *entity.Organization
	var project *entity.Project

	// find or prepare for creating project
	project = entity.FindProjectByOrganizationAndName(ctx, organizationName, projectName)
	if project == nil {
		organization = entity.FindOrganizationByName(ctx, organizationName)
		if organization == nil {
			return nil, gqlerror.Errorf("Organization %s not found", organizationName)
		}
		project = &entity.Project{
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
	var perms entity.ProjectPermissions
	secret := middleware.GetSecret(ctx)
	if create {
		orgPerms := secret.OrganizationPermissions(ctx, organization.OrganizationID)
		if !orgPerms.Create {
			return nil, gqlerror.Errorf("Not allowed to perform admin functions on organization %s", organizationName)
		}
		perms = entity.ProjectPermissions{
			View:   true,
			Create: true,
			Admin:  true,
		}
	} else {
		perms = secret.ProjectPermissions(ctx, project.ProjectID, false)
		if !perms.Admin {
			return nil, gqlerror.Errorf("Not allowed to perform admin functions on project %s/%s", organizationName, projectName)
		}
	}

	// default public to false upon creation
	if create && public == nil {
		falseVal := false
		public = &falseVal
	}

	err := project.StageWithUser(ctx, displayName, public, description, site, photoURL, secret.GetOwnerID(), perms)
	if err != nil {
		return nil, gqlerror.Errorf("Error staging project: %s", err.Error())
	}

	// on create, refetching project to include user
	if create {
		project = entity.FindProject(ctx, project.ProjectID)
		if project == nil {
			panic(fmt.Errorf("expected project with ID %s to exist", project.ProjectID.String()))
		}
	}

	return projectWithPermissions(project, perms), nil
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

func projectWithPermissions(p *entity.Project, perms entity.ProjectPermissions) *entity.Project {
	if perms.View || perms.Create || perms.Admin {
		p.Permissions = &entity.PermissionsUsersProjects{
			View:   perms.View,
			Create: perms.Create,
			Admin:  perms.Admin,
		}
	}
	return p
}
