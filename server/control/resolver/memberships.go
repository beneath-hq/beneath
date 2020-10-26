package resolver

import (
	"context"

	"gitlab.com/beneath-hq/beneath/models"
	"gitlab.com/beneath-hq/beneath/server/control/gql"
)

// ProjectMember returns the ProjectMemberResolver
func (r *Resolver) ProjectMember() gql.ProjectMemberResolver {
	return &projectMemberResolver{r}
}

type projectMemberResolver struct{ *Resolver }

func (r *projectMemberResolver) ProjectID(ctx context.Context, obj *models.ProjectMember) (string, error) {
	return obj.ProjectID.String(), nil
}

func (r *projectMemberResolver) UserID(ctx context.Context, obj *models.ProjectMember) (string, error) {
	return obj.UserID.String(), nil
}

// OrganizationMember returns the gql.OrganizationMemberResolver
func (r *Resolver) OrganizationMember() gql.OrganizationMemberResolver {
	return &organizationMemberResolver{r}
}

type organizationMemberResolver struct{ *Resolver }

func (r *organizationMemberResolver) OrganizationID(ctx context.Context, obj *models.OrganizationMember) (string, error) {
	return obj.OrganizationID.String(), nil
}

func (r *organizationMemberResolver) UserID(ctx context.Context, obj *models.OrganizationMember) (string, error) {
	return obj.UserID.String(), nil
}
