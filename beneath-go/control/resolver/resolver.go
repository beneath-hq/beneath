package resolver

import (
	"context"

	"github.com/beneath-core/beneath-go/control/gql"
	"github.com/beneath-core/beneath-go/control/model"
)

// THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

type Resolver struct{}

func (r *Resolver) Key() gql.KeyResolver {
	return &keyResolver{r}
}
func (r *Resolver) Mutation() gql.MutationResolver {
	return &mutationResolver{r}
}
func (r *Resolver) Project() gql.ProjectResolver {
	return &projectResolver{r}
}
func (r *Resolver) Query() gql.QueryResolver {
	return &queryResolver{r}
}
func (r *Resolver) Stream() gql.StreamResolver {
	return &streamResolver{r}
}
func (r *Resolver) Subscription() gql.SubscriptionResolver {
	return &subscriptionResolver{r}
}
func (r *Resolver) User() gql.UserResolver {
	return &userResolver{r}
}

type keyResolver struct{ *Resolver }

func (r *keyResolver) KeyID(ctx context.Context, obj *model.Key) (string, error) {
	panic("not implemented")
}

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) Empty(ctx context.Context) (*string, error) {
	panic("not implemented")
}
func (r *mutationResolver) IssueKey(ctx context.Context, userID *string, projectID *string, readonly bool, description *string) (*gql.NewKey, error) {
	panic("not implemented")
}
func (r *mutationResolver) RevokeKey(ctx context.Context, keyID string) (*bool, error) {
	panic("not implemented")
}
func (r *mutationResolver) CreateProject(ctx context.Context, name string, displayName string, site *string, description *string, photoURL *string) (*model.Project, error) {
	panic("not implemented")
}
func (r *mutationResolver) UpdateProject(ctx context.Context, projectID string, displayName *string, site *string, description *string, photoURL *string) (*model.Project, error) {
	panic("not implemented")
}
func (r *mutationResolver) AddUserToProject(ctx context.Context, email string, projectID string) (*model.User, error) {
	panic("not implemented")
}
func (r *mutationResolver) RemoveUserFromProject(ctx context.Context, userID string, projectID string) (*bool, error) {
	panic("not implemented")
}
func (r *mutationResolver) CreateExternalStream(ctx context.Context, projectID string, name string, description string, avroSchema string, batch bool, manual bool) (*model.Stream, error) {
	panic("not implemented")
}
func (r *mutationResolver) UpdateStream(ctx context.Context, streamID string, description *string, manual *bool) (*model.Stream, error) {
	panic("not implemented")
}
func (r *mutationResolver) UpdateMe(ctx context.Context, name *string, bio *string) (*gql.Me, error) {
	panic("not implemented")
}

type projectResolver struct{ *Resolver }

func (r *projectResolver) ProjectID(ctx context.Context, obj *model.Project) (string, error) {
	panic("not implemented")
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) Empty(ctx context.Context) (*string, error) {
	panic("not implemented")
}
func (r *queryResolver) Ping(ctx context.Context) (string, error) {
	panic("not implemented")
}
func (r *queryResolver) Keys(ctx context.Context, userID *string, projectID *string) ([]*model.Key, error) {
	panic("not implemented")
}
func (r *queryResolver) Project(ctx context.Context, name *string, projectID *string) (*model.Project, error) {
	panic("not implemented")
}
func (r *queryResolver) Stream(ctx context.Context, name string, projectName string) (*model.Stream, error) {
	panic("not implemented")
}
func (r *queryResolver) Me(ctx context.Context) (*gql.Me, error) {
	panic("not implemented")
}
func (r *queryResolver) User(ctx context.Context, userID string) (*model.User, error) {
	panic("not implemented")
}

type streamResolver struct{ *Resolver }

func (r *streamResolver) StreamID(ctx context.Context, obj *model.Stream) (string, error) {
	panic("not implemented")
}
func (r *streamResolver) SchemaType(ctx context.Context, obj *model.Stream) (*string, error) {
	panic("not implemented")
}

type subscriptionResolver struct{ *Resolver }

func (r *subscriptionResolver) Empty(ctx context.Context) (<-chan *string, error) {
	panic("not implemented")
}

type userResolver struct{ *Resolver }

func (r *userResolver) UserID(ctx context.Context, obj *model.User) (string, error) {
	panic("not implemented")
}
