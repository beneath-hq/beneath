package resolver

import (
	"context"

	"github.com/beneath-core/beneath-go/control/auth"
	"github.com/vektah/gqlparser/gqlerror"

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
	return obj.UserID.String(), nil
}

func (r *queryResolver) User(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	return model.FindUser(userID), nil
}

func (r *queryResolver) Me(ctx context.Context) (*gql.Me, error) {
	key := auth.GetKey(ctx)
	if !key.IsPersonal() {
		return nil, gqlerror.Errorf("Must be authenticated with a personal key to call 'me'")
	}

	user := model.FindUser(*key.UserID)
	return userToMe(user), nil
}

func (r *mutationResolver) UpdateMe(ctx context.Context, name *string, bio *string) (*gql.Me, error) {
	key := auth.GetKey(ctx)
	if !key.IsPersonal() {
		return nil, gqlerror.Errorf("Must be authenticated with a personal key to call 'updateMe'")
	}

	user := model.FindUser(*key.UserID)
	err := user.UpdateDescription(name, bio)
	if err != nil {
		return nil, err
	}

	return userToMe(user), nil
}

func userToMe(u *model.User) *gql.Me {
	if u == nil {
		return nil
	}
	return &gql.Me{
		UserID:    u.UserID.String(),
		User:      u,
		Email:     u.Email,
		UpdatedOn: u.UpdatedOn,
		Keys:      u.Keys,
	}
}
