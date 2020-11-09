package organization

import (
	"context"

	"github.com/go-pg/pg"
	uuid "github.com/satori/go.uuid"

	"gitlab.com/beneath-hq/beneath/bus"
	"gitlab.com/beneath-hq/beneath/infra/db"
	"gitlab.com/beneath-hq/beneath/models"
)

// Service encapsulates behavior for dealing with organizations
type Service struct {
	Bus *bus.Bus
	DB  db.DB
}

// New creates a new Service
func New(bus *bus.Bus, db db.DB) *Service {
	return &Service{
		Bus: bus,
		DB:  db,
	}
}

// FindOrganization finds a organization by ID
func (s *Service) FindOrganization(ctx context.Context, organizationID uuid.UUID) *models.Organization {
	organization := &models.Organization{
		OrganizationID: organizationID,
	}
	err := s.DB.GetDB(ctx).ModelContext(ctx, organization).
		WherePK().
		Column("organization.*", "User", "User.BillingOrganization", "Projects").
		Select()
	if !db.AssertFoundOne(err) {
		return nil
	}
	return organization
}

// FindOrganizationByName finds a organization by name
func (s *Service) FindOrganizationByName(ctx context.Context, name string) *models.Organization {
	organization := &models.Organization{}
	err := s.DB.GetDB(ctx).ModelContext(ctx, organization).
		Where("lower(organization.name) = lower(?)", name).
		Column(
			"organization.*",
			"User",
			"User.BillingOrganization",
			"Projects",
		).Select()
	if !db.AssertFoundOne(err) {
		return nil
	}
	return organization
}

// FindOrganizationByUserID finds an organization by its personal user ID (if set)
func (s *Service) FindOrganizationByUserID(ctx context.Context, userID uuid.UUID) *models.Organization {
	organization := &models.Organization{}
	err := s.DB.GetDB(ctx).ModelContext(ctx, organization).
		Where("organization.user_id = ?", userID).
		Column(
			"organization.*",
			"User",
			"User.BillingOrganization",
			"Projects",
		).Select()
	if !db.AssertFoundOne(err) {
		return nil
	}
	return organization
}

// FindAllOrganizations returns all active organizations
func (s *Service) FindAllOrganizations(ctx context.Context) []*models.Organization {
	var organizations []*models.Organization
	err := s.DB.GetDB(ctx).ModelContext(ctx, &organizations).
		Column("organization.*").
		Select()
	if err != nil {
		panic(err)
	}
	return organizations
}

// FindOrganizationMembers finds useful information about the organization's members, represented in
// the structure of OrganizationMember (includes both people with and without billing affiliation)
func (s *Service) FindOrganizationMembers(ctx context.Context, organizationID uuid.UUID) ([]*models.OrganizationMember, error) {
	var result []*models.OrganizationMember
	_, err := s.DB.GetDB(ctx).QueryContext(ctx, &result, `
		select
			p.organization_id,
			p.user_id,
			u.billing_organization_id,
			o.name,
			o.display_name,
			o.photo_url,
			p.view,
			p."create",
			p.admin,
			u.quota_epoch,
			u.read_quota,
			u.write_quota,
			u.scan_quota
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
