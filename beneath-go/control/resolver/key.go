package resolver

import (
	"context"

	"github.com/beneath-core/beneath-go/control/auth"
	"github.com/beneath-core/beneath-go/control/gql"
	"github.com/beneath-core/beneath-go/control/model"
	uuid "github.com/satori/go.uuid"
	"github.com/vektah/gqlparser/gqlerror"
)

// Key returns the gql.KeyResolver
func (r *Resolver) Key() gql.KeyResolver {
	return &keyResolver{r}
}

type keyResolver struct{ *Resolver }

func (r *keyResolver) KeyID(ctx context.Context, obj *model.Key) (string, error) {
	return obj.KeyID.String(), nil
}

func (r *keyResolver) Role(ctx context.Context, obj *model.Key) (string, error) {
	return string(obj.Role), nil
}

func (r *queryResolver) KeysForUser(ctx context.Context, userID uuid.UUID) ([]*model.Key, error) {
	key := auth.GetKey(ctx)
	if !key.IsPersonal() {
		return nil, gqlerror.Errorf("Must be authenticated with a personal key")
	}

	if userID != *key.UserID {
		return nil, gqlerror.Errorf("Not allowed to read user's keys")
	}

	return model.FindUserKeys(userID), nil
}

func (r *queryResolver) KeysForProject(ctx context.Context, projectID uuid.UUID) ([]*model.Key, error) {
	key := auth.GetKey(ctx)
	if !key.EditsProject(projectID) {
		return nil, gqlerror.Errorf("Not allowed to read project keys")
	}

	return model.FindProjectKeys(projectID), nil
}

func (r *mutationResolver) IssueUserKey(ctx context.Context, readonly bool, description string) (*gql.NewKey, error) {
	authKey := auth.GetKey(ctx)
	if !authKey.IsPersonal() {
		return nil, gqlerror.Errorf("Must be authenticated with a personal key")
	}

	role := model.KeyRoleReadWrite
	if readonly {
		role = model.KeyRoleReadonly
	}

	key, err := model.CreateUserKey(*authKey.UserID, role, description)
	if err != nil {
		return nil, gqlerror.Errorf(err.Error())
	}

	return &gql.NewKey{
		Key:       key,
		KeyString: key.KeyString,
	}, nil
}

func (r *mutationResolver) IssueProjectKey(ctx context.Context, projectID uuid.UUID, readonly bool, description string) (*gql.NewKey, error) {
	authKey := auth.GetKey(ctx)
	if !authKey.EditsProject(projectID) {
		return nil, gqlerror.Errorf("Not allowed to edit project")
	}

	role := model.KeyRoleReadWrite
	if readonly {
		role = model.KeyRoleReadonly
	}

	key, err := model.CreateProjectKey(projectID, role, description)
	if err != nil {
		return nil, gqlerror.Errorf(err.Error())
	}

	return &gql.NewKey{
		Key:       key,
		KeyString: key.KeyString,
	}, nil
}

func (r *mutationResolver) RevokeKey(ctx context.Context, keyID uuid.UUID) (bool, error) {
	key := model.FindKey(keyID)
	if key == nil {
		return false, gqlerror.Errorf("Key not found")
	}

	authKey := auth.GetKey(ctx)
	if key.ProjectID != nil && !authKey.EditsProject(*key.ProjectID) {
		return false, gqlerror.Errorf("Not allowed to edit key")
	} else if key.UserID != nil && *authKey.UserID != *key.UserID {
		return false, gqlerror.Errorf("Not allowed to edit key")
	}

	key.Revoke()
	return true, nil
}
