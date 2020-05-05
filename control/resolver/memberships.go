package resolver

import (
	"context"

	"gitlab.com/beneath-hq/beneath/control/entity"
	"gitlab.com/beneath-hq/beneath/control/gql"
)

// ProjectMember returns the ProjectMemberResolver
func (r *Resolver) ProjectMember() gql.ProjectMemberResolver {
	return &projectMemberResolver{r}
}

type projectMemberResolver struct{ *Resolver }

func (r *projectMemberResolver) UserID(ctx context.Context, obj *entity.ProjectMember) (string, error) {
	return obj.UserID.String(), nil
}

// OrganizationMember returns the gql.OrganizationMemberResolver
func (r *Resolver) OrganizationMember() gql.OrganizationMemberResolver {
	return &organizationMemberResolver{r}
}

type organizationMemberResolver struct{ *Resolver }

func (r *organizationMemberResolver) UserID(ctx context.Context, obj *entity.OrganizationMember) (string, error) {
	return obj.UserID.String(), nil
}
