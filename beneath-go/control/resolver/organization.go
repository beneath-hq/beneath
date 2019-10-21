package resolver

import (
	"context"

	"github.com/beneath-core/beneath-go/control/entity"
	"github.com/beneath-core/beneath-go/control/gql"
	"github.com/beneath-core/beneath-go/core/middleware"
	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/gqlerror"
)

// Organization returns the gql.OrganizationResolver
func (r *Resolver) Organization() gql.OrganizationResolver {
	return &organizationResolver{r}
}

type organizationResolver struct{ *Resolver }

func (r *organizationResolver) OrganizationID(ctx context.Context, obj *entity.Organization) (string, error) {
	return obj.OrganizationID.String(), nil
}

func (r *queryResolver) OrganizationByName(ctx context.Context, name string) (*entity.Organization, error) {
	organization := entity.FindOrganizationByName(ctx, name)
	if organization == nil {
		return nil, gqlerror.Errorf("Organization %s not found", name)
	}

	secret := middleware.GetSecret(ctx)
	perms := secret.OrganizationPermissions(ctx, organization.OrganizationID)
	if !perms.View {
		return nil, gqlerror.Errorf("Not allowed to view organization %s", name)
	}

	return organization, nil
}

func (r *mutationResolver) CreateOrganization(ctx context.Context, name string) (*entity.Organization, error) {
	secret := middleware.GetSecret(ctx)
	if !secret.IsUser() {
		return nil, gqlerror.Errorf("Not allowed to create organization")
	}

	org, err := entity.CreateOrganizationWithUser(ctx, name, *secret.UserID)
	if err != nil {
		return nil, err
	}

	return org, nil
}

func (r *mutationResolver) AddUserToOrganization(ctx context.Context, username string, organizationID uuid.UUID, view bool, admin bool) (*entity.User, error) {
	organization := entity.FindOrganization(ctx, organizationID)
	if organization == nil {
		return nil, gqlerror.Errorf("Organization %s not found", organizationID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := secret.OrganizationPermissions(ctx, organizationID)
	if !perms.Admin {
		return nil, gqlerror.Errorf("Not allowed to perform admin functions on organization %s", organizationID.String())
	}

	user := entity.FindUserByUsername(ctx, username)
	if user == nil {
		return nil, gqlerror.Errorf("No user found with that username")
	}

	err := organization.AddUser(ctx, user.UserID, view, admin)
	if err != nil {
		return nil, gqlerror.Errorf(err.Error())
	}

	return user, nil
}

func (r *mutationResolver) RemoveUserFromOrganization(ctx context.Context, userID uuid.UUID, organizationID uuid.UUID) (bool, error) {
	organization := entity.FindOrganization(ctx, organizationID)
	if organization == nil {
		return false, gqlerror.Errorf("Organization %s not found", organizationID.String())
	}

	secret := middleware.GetSecret(ctx)
	perms := secret.OrganizationPermissions(ctx, organizationID)
	if !perms.Admin {
		return false, gqlerror.Errorf("Not allowed to perform admin functions in organization %s", organizationID.String())
	}

	if len(organization.Users) < 2 {
		return false, gqlerror.Errorf("Can't remove last member of organization")
	}

	err := organization.RemoveUser(ctx, userID)
	if err != nil {
		return false, gqlerror.Errorf(err.Error())
	}

	return true, nil
}
