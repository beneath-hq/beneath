package resolver

import (
	"context"

	"github.com/beneath-core/beneath-go/control/gql"
	"github.com/beneath-core/beneath-go/control/model"
	uuid "github.com/satori/go.uuid"
)

// User returns the gql.UserResolver
func (r *Resolver) User() gql.UserResolver {
	return &userResolver{r}
}

type userResolver struct{ *Resolver }

func (r *userResolver) UserID(ctx context.Context, obj *model.User) (string, error) {
	panic("not implemented")
}

func (r *queryResolver) User(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	panic("not implemented")
}

func (r *queryResolver) Me(ctx context.Context) (*gql.Me, error) {
	panic("not implemented")
}

func (r *mutationResolver) UpdateMe(ctx context.Context, name *string, bio *string) (*gql.Me, error) {
	panic("not implemented")
}
