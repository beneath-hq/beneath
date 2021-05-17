package organization

import (
	"context"

	"github.com/beneath-hq/beneath/infra/db"
	"github.com/beneath-hq/beneath/models"
	uuid "github.com/satori/go.uuid"
)

// An invite procedure is required for organizations since there's mutual costs and benefits: a) The inviter
// proposes to cover the invitees bills, but b) the invitee is subjected to quotas set by
// from the inviter (as well as more intrusive oversight).
// Don't confuse changing billing organizations with simply granted organization permissions to someone,
// which doesn't imply a change in billing or quotas.

// FindOrganizationInvite finds an existing invitation
func (s *Service) FindOrganizationInvite(ctx context.Context, organizationID uuid.UUID, userID uuid.UUID) *models.OrganizationInvite {
	invite := &models.OrganizationInvite{}
	err := s.DB.GetDB(ctx).ModelContext(ctx, invite).
		Column("organization_invite.*", "Organization", "User", "User.BillingOrganization").
		Where("organization_invite.organization_id = ?", organizationID).
		Where("organization_invite.user_id = ?", userID).
		Select()
	if !db.AssertFoundOne(err) {
		return nil
	}
	return invite
}

// CreateOrUpdateInvite creates or updates the organization invite
func (s *Service) CreateOrUpdateInvite(ctx context.Context, organizationID uuid.UUID, userID uuid.UUID, view, create, admin bool) error {
	invite := &models.OrganizationInvite{
		OrganizationID: organizationID,
		UserID:         userID,
		View:           view,
		Create:         create,
		Admin:          admin,
	}

	err := invite.Validate()
	if err != nil {
		return err
	}

	// build upsert
	q := s.DB.GetDB(ctx).ModelContext(ctx, invite).OnConflict("(organization_id, user_id) DO UPDATE")
	q.Set("view = EXCLUDED.view")
	q.Set(`"create" = EXCLUDED."create"`)
	q.Set("admin = EXCLUDED.admin")

	// run upsert
	_, err = q.Insert()
	if err != nil {
		return err
	}

	return nil
}
