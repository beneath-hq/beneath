package resolver

import (
	"context"

	"github.com/vektah/gqlparser/gqlerror"

	"github.com/beneath-core/beneath-go/control/entity"
	"github.com/beneath-core/beneath-go/control/gql"
	"github.com/beneath-core/beneath-go/core/middleware"
	"github.com/beneath-core/beneath-go/metrics"
	uuid "github.com/satori/go.uuid"
)

// User returns the gql.UserResolver
func (r *Resolver) User() gql.UserResolver {
	return &userResolver{r}
}

type userResolver struct{ *Resolver }

func (r *userResolver) UserID(ctx context.Context, obj *entity.User) (string, error) {
	return obj.UserID.String(), nil
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

func (r *queryResolver) Me(ctx context.Context) (*gql.Me, error) {
	secret := middleware.GetSecret(ctx)
	if !secret.IsUser() {
		return nil, MakeUnauthenticatedError("Must be authenticated with a personal key to call 'Me'")
	}

	user := entity.FindUser(ctx, secret.GetOwnerID())
	return userToMe(ctx, user), nil
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
		UserID:       u.UserID.String(),
		User:         u,
		Email:        u.Email,
		ReadUsage:    int(usage.ReadBytes),
		ReadQuota:    int(u.ReadQuota),
		WriteUsage:   int(usage.WriteBytes),
		WriteQuota:   int(u.WriteQuota),
		UpdatedOn:    u.UpdatedOn,
		Organization: u.Organization,
	}
}

func (r *mutationResolver) JoinOrganization(ctx context.Context, organizationName string) (*gql.Me, error) {
	secret := middleware.GetSecret(ctx)
	if !secret.IsUser() {
		return nil, MakeUnauthenticatedError("Must be authenticated with a personal key")
	}

	organization := entity.FindOrganizationByName(ctx, organizationName)
	if organization == nil {
		return nil, gqlerror.Errorf("Organization %s not found", organizationName)
	}

	perms := secret.OrganizationPermissions(ctx, organization.OrganizationID)
	if !perms.View {
		return nil, gqlerror.Errorf("You don't have permission to join organization %s", organizationName)
	}

	user := entity.FindUser(ctx, secret.GetOwnerID())

	if user.Organization.Name == organizationName {
		return nil, gqlerror.Errorf("You are already a member of organization %s", organizationName)
	}

	user, err := user.JoinOrganization(ctx, organization.OrganizationID)
	if err != nil {
		return nil, err
	}

	return userToMe(ctx, user), nil
}
