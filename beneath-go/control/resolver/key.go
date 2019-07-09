package resolver

import (
	"context"

	"github.com/beneath-core/beneath-go/control/gql"
	"github.com/beneath-core/beneath-go/control/model"
	uuid "github.com/satori/go.uuid"
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
	panic("not implemented")
}
func (r *queryResolver) KeysForProject(ctx context.Context, projectID uuid.UUID) ([]*model.Key, error) {
	panic("not implemented")
}
func (r *mutationResolver) IssueUserKey(ctx context.Context, userID uuid.UUID, readonly bool, description string) (*gql.NewKey, error) {
	panic("not implemented")
}
func (r *mutationResolver) IssueProjectKey(ctx context.Context, projectID uuid.UUID, readonly bool, description string) (*gql.NewKey, error) {
	panic("not implemented")
}
func (r *mutationResolver) RevokeKey(ctx context.Context, keyID uuid.UUID) (bool, error) {
	panic("not implemented")
}
