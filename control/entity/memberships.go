package entity

import (
	"context"

	"github.com/go-pg/pg/v9"
	uuid "github.com/satori/go.uuid"

	"gitlab.com/beneath-hq/beneath/internal/hub"
)

// ProjectMember is a convenience representation of project membership
type ProjectMember struct {
	ProjectID   uuid.UUID
	UserID      uuid.UUID
	Name        string
	DisplayName string
	View        bool
	Create      bool
	Admin       bool
}

// OrganizationMember is a convenience representation of organization membership
type OrganizationMember struct {
	OrganizationID        uuid.UUID
	UserID                uuid.UUID
	BillingOrganizationID uuid.UUID
	Name                  string
	DisplayName           string
	View                  bool
	Create                bool
	Admin                 bool
	ReadQuota             *int
	WriteQuota            *int
}

// FindProjectMembers finds useful information about the project's members, represented in
// the structure of ProjectMember
func FindProjectMembers(ctx context.Context, projectID uuid.UUID) ([]*ProjectMember, error) {
	var result []*ProjectMember
	_, err := hub.DB.QueryContext(ctx, &result, `
		select
			p.project_id,
			p.user_id,
			o.name,
			o.display_name,
			p.view,
			p."create",
			p.admin
		from permissions_users_projects p
		join organizations o on p.user_id = o.user_id
		where p.project_id = ?
	`, projectID)
	if err != nil && err != pg.ErrNoRows {
		return nil, err
	}
	return result, nil
}

// FindOrganizationMembers finds useful information about the organization's members, represented in
// the structure of OrganizationMember (includes both people with and without billing affiliation)
func FindOrganizationMembers(ctx context.Context, organizationID uuid.UUID) ([]*OrganizationMember, error) {
	var result []*OrganizationMember
	_, err := hub.DB.QueryContext(ctx, &result, `
		select
			p.organization_id,
			p.user_id,
			u.billing_organization_id,
			o.name,
			o.display_name,
			p.view,
			p."create",
			p.admin,
			u.read_quota,
			u.write_quota
		from permissions_users_organizations p
		join organizations o on p.user_id = o.user_id
		join users u on p.user_id = u.user_id
		where p.organization_id = ?
	`, organizationID)
	if err != nil && err != pg.ErrNoRows {
		return nil, err
	}
	return result, nil
}
