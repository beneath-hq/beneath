package resolver

import (
	"context"

	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/gqlerror"

	"gitlab.com/beneath-hq/beneath/control/entity"
	"gitlab.com/beneath-hq/beneath/control/gql"
	"gitlab.com/beneath-hq/beneath/internal/middleware"
)

// PrivateUser returns the gql.UserResolver
func (r *Resolver) PrivateUser() gql.PrivateUserResolver {
	return &privateUserResolver{r}
}

type privateUserResolver struct{ *Resolver }

func (r *privateUserResolver) UserID(ctx context.Context, obj *entity.User) (string, error) {
	return obj.UserID.String(), nil
}

func (r *mutationResolver) UpdateUserQuotas(ctx context.Context, userID uuid.UUID, readQuota *int, writeQuota *int) (*entity.User, error) {
	user := entity.FindUser(ctx, userID)
	if user == nil {
		return nil, gqlerror.Errorf("User %s not found", userID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := secret.OrganizationPermissions(ctx, user.BillingOrganizationID)
	if !perms.Admin {
		return nil, gqlerror.Errorf("Not allowed to perform admin functions in organization %s", user.BillingOrganizationID.String())
	}

	// TODO: invalidate cached quotas

	err := user.UpdateQuotas(ctx, IntToInt64(readQuota), IntToInt64(writeQuota))
	if err != nil {
		return nil, gqlerror.Errorf("Error updating quotas: %s", err.Error())
	}

	return user, nil
}

func (r *mutationResolver) UpdateUserProjectPermissions(ctx context.Context, userID uuid.UUID, projectID uuid.UUID, view *bool, create *bool, admin *bool) (*entity.PermissionsUsersProjects, error) {
	secret := middleware.GetSecret(ctx)
	perms := secret.ProjectPermissions(ctx, projectID, false)
	if !perms.Admin {
		return nil, gqlerror.Errorf("Not allowed to perform admin functions in project %s", projectID.String())
	}

	pup := entity.FindPermissionsUsersProjects(ctx, userID, projectID)
	if pup == nil {
		pup = &entity.PermissionsUsersProjects{
			UserID:    userID,
			ProjectID: projectID,
		}
	}

	err := pup.Update(ctx, view, create, admin)
	if err != nil {
		return nil, err
	}

	return pup, nil
}

func (r *mutationResolver) UpdateUserOrganizationPermissions(ctx context.Context, userID uuid.UUID, organizationID uuid.UUID, view *bool, create *bool, admin *bool) (*entity.PermissionsUsersOrganizations, error) {
	secret := middleware.GetSecret(ctx)
	perms := secret.OrganizationPermissions(ctx, organizationID)
	if !perms.Admin {
		return nil, gqlerror.Errorf("Not allowed to perform admin functions in organization %s", organizationID.String())
	}

	organization := entity.FindOrganization(ctx, organizationID)
	if organization == nil {
		return nil, gqlerror.Errorf("Organization not found")
	}

	if !organization.IsMulti() {
		return nil, gqlerror.Errorf("Cannot edit permissions of personal organization")
	}

	puo := entity.FindPermissionsUsersOrganizations(ctx, userID, organizationID)
	if puo == nil {
		puo = &entity.PermissionsUsersOrganizations{
			UserID:         userID,
			OrganizationID: organizationID,
		}
	}

	err := puo.Update(ctx, view, create, admin)
	if err != nil {
		return nil, err
	}

	return puo, nil
}
