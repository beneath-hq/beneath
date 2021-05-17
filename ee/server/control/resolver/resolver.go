package resolver

import (
	"context"

	"github.com/beneath-hq/beneath/ee/server/control/gql"
	"github.com/beneath-hq/beneath/ee/services/billing"
	"github.com/beneath-hq/beneath/services/organization"
	"github.com/beneath-hq/beneath/services/permissions"
)

// Resolver implements gql.ResolverRoot
type Resolver struct {
	Billing       *billing.Service
	Permissions   *permissions.Service
	Organizations *organization.Service
}

// Query returns the gql.QueryResolver
func (r *Resolver) Query() gql.QueryResolver {
	return &queryResolver{r}
}

type queryResolver struct{ *Resolver }

// Empty is part of gql.QueryResolver
func (r *queryResolver) Empty(ctx context.Context) (*string, error) {
	panic("not implemented")
}

// Mutation returns the gql.MutationResolver
func (r *Resolver) Mutation() gql.MutationResolver {
	return &mutationResolver{r}
}

type mutationResolver struct{ *Resolver }

// Empty is part of gql.MutationResolver
func (r *mutationResolver) Empty(ctx context.Context) (*string, error) {
	panic("not implemented")
}
