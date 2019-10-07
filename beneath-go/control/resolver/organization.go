package resolver

import (
	"context"

	"github.com/beneath-core/beneath-go/control/entity"
	"github.com/beneath-core/beneath-go/control/gql"
	"github.com/beneath-core/beneath-go/core/middleware"
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

func (r *mutationResolver) CreateOrganization(ctx context.Context, name string) (*entity.Organization, error) {
	secret := middleware.GetSecret(ctx)
	if !secret.IsUser() {
		return nil, gqlerror.Errorf("Not allowed to create organization")
	}

	organization := &entity.Organization{
		Name: name,
	}

	err := organization.Create(ctx, name)
	if err != nil {
		return nil, err
	}

	return organization, nil
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
