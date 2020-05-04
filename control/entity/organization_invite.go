package entity

import (
	"context"
	"time"

	uuid "github.com/satori/go.uuid"
	"gitlab.com/beneath-hq/beneath/internal/hub"
)

// OrganizationInvite represents invites to use an organization as your billing organization.
// An invite procedure is required since there's mutual costs and benefits: a) The inviter
// proposes to cover the invitees bills, but b) the invitee is subjected to quotas set by
// from the inviter (as well as more intrusive oversight).
// Don't confuse changing billing organizations with simply granted organization permissions to someone,
// which doesn't imply a change in billing or quotas.
type OrganizationInvite struct {
	OrganizationInviteID uuid.UUID `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	OrganizationID       uuid.UUID `sql:"on_delete:CASCADE,type:uuid"`
	Organization         *Organization
	UserID               uuid.UUID `sql:"on_delete:CASCADE,type:uuid"`
	User                 *User
	CreatedOn            time.Time `sql:",default:now()"`
	UpdatedOn            time.Time `sql:",default:now()"`
	View                 bool      `sql:",notnull"`
	Create               bool      `sql:",notnull"`
	Admin                bool      `sql:",notnull"`
}

// FindOrganizationInvite finds an existing invitation
func FindOrganizationInvite(ctx context.Context, organizationID uuid.UUID, userID uuid.UUID) *OrganizationInvite {
	invite := &OrganizationInvite{}
	err := hub.DB.ModelContext(ctx, invite).
		Column("invite.*", "Organization", "User", "User.BillingOrganization").
		Where("organization_id = ?", organizationID).
		Where("user_id = ?", userID).
		Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return invite
}

// Save creates or updates the organization invite
func (i *OrganizationInvite) Save(ctx context.Context) error {
	// validate
	err := GetValidator().Struct(i)
	if err != nil {
		return err
	}

	// build upsert
	q := hub.DB.ModelContext(ctx, i).OnConflict("(organization_id, user_id) DO UPDATE")
	q.Set("view = EXCLUDED.view")
	q.Set("create = EXCLUDED.create")
	q.Set("admin = EXCLUDED.admin")

	// run upsert
	_, err = q.Insert()
	if err != nil {
		return err
	}

	return nil
}
