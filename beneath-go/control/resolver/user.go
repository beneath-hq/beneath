package resolver

import (
	"context"

	"github.com/vektah/gqlparser/gqlerror"

	"github.com/beneath-core/beneath-go/control/entity"
	"github.com/beneath-core/beneath-go/control/gql"
	"github.com/beneath-core/beneath-go/core/middleware"
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

func (r *queryResolver) Me(ctx context.Context) (*gql.Me, error) {
	secret := middleware.GetSecret(ctx)
	if !secret.IsPersonal() {
		return nil, MakeUnauthenticatedError("Must be authenticated with a personal key to call 'Me'")
	}

	user := entity.FindUser(ctx, *secret.UserID)
	return userToMe(user), nil
}

func (r *mutationResolver) UpdateMe(ctx context.Context, name *string, bio *string) (*gql.Me, error) {
	secret := middleware.GetSecret(ctx)
	if !secret.IsPersonal() {
		return nil, MakeUnauthenticatedError("Must be authenticated with a personal key to call 'updateMe'")
	}

	user := entity.FindUser(ctx, *secret.UserID)
	err := user.UpdateDescription(ctx, name, bio)
	if err != nil {
		return nil, err
	}

	return userToMe(user), nil
}

func userToMe(u *entity.User) *gql.Me {
	if u == nil {
		return nil
	}
	return &gql.Me{
		UserID:    u.UserID.String(),
		User:      u,
		Email:     u.Email,
		UpdatedOn: u.UpdatedOn,
	}
}
