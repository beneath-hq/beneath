package resolver

import (
	"context"

	"github.com/beneath-core/beneath-go/control/gql"
	"github.com/beneath-core/beneath-go/control/model"
	uuid "github.com/satori/go.uuid"
)

// Project returns the gql.ProjectResolver
func (r *Resolver) Project() gql.ProjectResolver {
	return &projectResolver{r}
}

type projectResolver struct{ *Resolver }

func (r *projectResolver) ProjectID(ctx context.Context, obj *model.Project) (string, error) {
	return obj.ProjectID.String(), nil
}

func (r *queryResolver) ProjectByName(ctx context.Context, name string) (*model.Project, error) {
	panic("not implemented")
}

func (r *queryResolver) ProjectByID(ctx context.Context, projectID uuid.UUID) (*model.Project, error) {
	panic("not implemented")
}

func (r *mutationResolver) CreateProject(ctx context.Context, name string, displayName string, site *string, description *string, photoURL *string) (*model.Project, error) {
	panic("not implemented")
}

func (r *mutationResolver) UpdateProject(ctx context.Context, projectID uuid.UUID, displayName *string, site *string, description *string, photoURL *string) (*model.Project, error) {
	panic("not implemented")
}

func (r *mutationResolver) AddUserToProject(ctx context.Context, email string, projectID uuid.UUID) (*model.User, error) {
	panic("not implemented")
}

func (r *mutationResolver) RemoveUserFromProject(ctx context.Context, userID uuid.UUID, projectID uuid.UUID) (bool, error) {
	panic("not implemented")
}
