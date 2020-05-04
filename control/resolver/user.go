package resolver

import (
	"context"

	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/gqlerror"

	"gitlab.com/beneath-hq/beneath/control/entity"
	"gitlab.com/beneath-hq/beneath/control/gql"
	"gitlab.com/beneath-hq/beneath/internal/metrics"
	"gitlab.com/beneath-hq/beneath/internal/middleware"
)

// User returns the gql.UserResolver
func (r *Resolver) User() gql.UserResolver {
	return &userResolver{r}
}

type userResolver struct{ *Resolver }

func (r *userResolver) UserID(ctx context.Context, obj *entity.User) (string, error) {
	return obj.UserID.String(), nil
}

func (r *queryResolver) Me(ctx context.Context) (*gql.Me, error) {
	secret := middleware.GetSecret(ctx)
	if !secret.IsUser() {
		return nil, MakeUnauthenticatedError("Must be authenticated with a personal key to call 'Me'")
	}

	user := entity.FindUser(ctx, secret.GetOwnerID())
	return userToMe(ctx, user), nil
}

func (r *queryResolver) User(ctx context.Context, userID uuid.UUID) (*entity.User, error) {
	user := entity.FindUser(ctx, userID)
	if user == nil {
		return nil, gqlerror.Errorf("User %s not found", userID.String())
	}
	return user, nil
}

func (r *queryResolver) UserByUsername(ctx context.Context, username string) (*entity.User, error) {
	user := entity.FindUserByUsername(ctx, username)
	if user == nil {
		return nil, gqlerror.Errorf("User %s not found", username)
	}
	return user, nil
}

func (r *mutationResolver) UpdateMe(ctx context.Context, username *string, name *string, bio *string, photoURL *string) (*gql.Me, error) {
	secret := middleware.GetSecret(ctx)
	if !secret.IsUser() {
		return nil, MakeUnauthenticatedError("Must be authenticated with a personal key to call 'updateMe'")
	}

	user := entity.FindUser(ctx, secret.GetOwnerID())
	err := user.UpdateDescription(ctx, username, name, bio, photoURL)
	if err != nil {
		return nil, err
	}

	return userToMe(ctx, user), nil
}

func userToMe(ctx context.Context, u *entity.User) *gql.Me {
	if u == nil {
		return nil
	}

	usage := metrics.GetCurrentUsage(ctx, u.UserID)

	return &gql.Me{
		UserID:               u.UserID.String(),
		User:                 u,
		Email:                u.Email,
		ReadUsage:            int(usage.ReadBytes),
		ReadQuota:            int(u.ReadQuota),
		WriteUsage:           int(usage.WriteBytes),
		WriteQuota:           int(u.WriteQuota),
		UpdatedOn:            u.UpdatedOn,
		PersonalOrganization: u.PersonalOrganization,
		BillingOrganization:  u.BillingOrganization,
	}
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
