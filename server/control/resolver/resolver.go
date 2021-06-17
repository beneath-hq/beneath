package resolver

import (
	"context"

	"github.com/beneath-hq/beneath/server/control/gql"
	"github.com/beneath-hq/beneath/services/organization"
	"github.com/beneath-hq/beneath/services/permissions"
	"github.com/beneath-hq/beneath/services/project"
	"github.com/beneath-hq/beneath/services/secret"
	"github.com/beneath-hq/beneath/services/service"
	"github.com/beneath-hq/beneath/services/table"
	"github.com/beneath-hq/beneath/services/usage"
	"github.com/beneath-hq/beneath/services/user"
)

// Resolver implements gql.ResolverRoot
type Resolver struct {
	Usage         *usage.Service
	Organizations *organization.Service
	Permissions   *permissions.Service
	Projects      *project.Service
	Secrets       *secret.Service
	Services      *service.Service
	Tables        *table.Service
	Users         *user.Service
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

// Ping is part of gql.QueryResolver
func (r *queryResolver) Ping(ctx context.Context) (string, error) {
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

// Subscription returns the gql.SubscriptionResolver
func (r *Resolver) Subscription() gql.SubscriptionResolver {
	return &subscriptionResolver{r}
}

type subscriptionResolver struct{ *Resolver }

// Empty is part of gql.SubscriptionResolver
func (r *subscriptionResolver) Empty(ctx context.Context) (<-chan *string, error) {
	panic("not implemented")
}
