package entity

import (
	"context"
	"regexp"
	"time"

	uuid "github.com/satori/go.uuid"
	"gopkg.in/go-playground/validator.v9"

	"github.com/beneath-core/beneath-go/db"
)

// Organization represents the entity that manages billing on behalf of its users
type Organization struct {
	OrganizationID uuid.UUID `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	Name           string    `sql:",unique,notnull",validate:"required,gte=1,lte=40"`
	Personal       bool      `sql:",notnull"`
	Active         bool      `sql:",notnull"`
	CreatedOn      time.Time `sql:",default:now()"`
	UpdatedOn      time.Time `sql:",default:now()"`
	DeletedOn      time.Time
	Services       []*Service
	Users          []*User
}

var (
	// used for validation
	organizationNameRegex *regexp.Regexp
)

func init() {
	// configure validation
	organizationNameRegex = regexp.MustCompile("^[_a-z][_a-z0-9]*$")
	GetValidator().RegisterStructValidation(organizationValidation, Organization{})
}

// custom stream validation
func organizationValidation(sl validator.StructLevel) {
	s := sl.Current().Interface().(Organization)

	if !organizationNameRegex.MatchString(s.Name) {
		sl.ReportError(s.Name, "Name", "", "alphanumericorunderscore", "")
	}
}

// FindOrganization finds a organization by ID
func FindOrganization(ctx context.Context, organizationID uuid.UUID) *Organization {
	organization := &Organization{
		OrganizationID: organizationID,
	}
	err := db.DB.ModelContext(ctx, organization).WherePK().Column("organization.*", "Services", "Users").Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return organization
}

// FindOrganizationByName finds a organization by name
func FindOrganizationByName(ctx context.Context, name string) *Organization {
	organization := &Organization{}
	err := db.DB.ModelContext(ctx, organization).
		Where("lower(name) = lower(?)", name).
		Column("organization.*", "Services", "Users").
		Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return organization
}

// FindActiveOrganizations returns all active organizations
func FindActiveOrganizations(ctx context.Context) []*Organization {
	var organizations []*Organization
	err := db.DB.ModelContext(ctx, &organizations).
		Where("active = true").
		Column("organization.*").
		Select()
	if err != nil {
		panic(err)
	}
	return organizations
}

// UpdatePersonalStatus updates an organization's "personal" status
// this is called when a user changes plans to/from an Enterprise plan
func (o *Organization) UpdatePersonalStatus(ctx context.Context, personal bool) error {
	org := &Organization{
		OrganizationID: o.OrganizationID,
		Personal:       personal,
	}

	_, err := db.DB.ModelContext(ctx, org).
		Column("personal").
		WherePK().
		Update()
	if err != nil {
		return err
	}

	return nil
}

// UpdateActiveStatus updates an organization's "active" status
func (o *Organization) UpdateActiveStatus(ctx context.Context, active bool) error {
	org := &Organization{
		OrganizationID: o.OrganizationID,
		Active:         active,
	}

	_, err := db.DB.ModelContext(ctx, org).
		Column("active").
		WherePK().
		Update()
	if err != nil {
		return err
	}

	return nil
}

// AddUser makes user a member of organization
// note: in practice, this function really means "invite user"; as the user won't officially be a part of the organization until they "join organization" (i.e. accept the invitation)
// TODO: clear secret cache!
func (o *Organization) AddUser(ctx context.Context, userID uuid.UUID, view bool, admin bool) error {
	perms := &PermissionsUsersOrganizations{
		UserID:         userID,
		OrganizationID: o.OrganizationID,
		View:           view,
		Admin:          admin,
	}

	_, err := db.DB.ModelContext(ctx, perms).OnConflict("(user_id, organization_id) DO UPDATE").Insert()
	if err != nil {
		return err
	}

	// done
	return nil
}

// ChangeUserPermissions changes a user's permissions within the organization
func (o *Organization) ChangeUserPermissions(ctx context.Context, userID uuid.UUID, view bool, admin bool) error {
	// TODO: convert permissions.Update (below) to this function
	return nil
}

// Update changes a user's permissions within the organization
// func (p *PermissionsUsersOrganizations) Update(ctx context.Context, view *bool, admin *bool) (*PermissionsUsersOrganizations, error) {
// 	// Q: this won't fuck up the object "p" that I have, right?
//  // A: No, it just invalidates it from the cache so other processes won't be served it
//  // A2: But, we should actually move it to after the DB update for consistency
// 	getUserOrganizationPermissionsCache().Clear(ctx, p.UserID, p.OrganizationID)

// 	if view != nil {
// 		p.View = *view
// 	}
// 	if admin != nil {
// 		p.Admin = *admin
// 	}
// 	_, err := db.DB.ModelContext(ctx, p).Column("view", "admin").WherePK().Update()
// 	return p, err
// }

// ChangeUserQuotas allows an admin to configure a member's quotas
func (o *Organization) ChangeUserQuotas(ctx context.Context, userID uuid.UUID, readQuota int64, writeQuota int64) error {
	// TODO: convert user.UpdateQuotas (below) to this function
	return nil
}

// UpdateQuotas updates user's quotas
// Q: do I have to invalidate the getSecretCache?
// func (u *User) UpdateQuotas(ctx context.Context, readQuota *int, writeQuota *int) (*User, error) {
// 	if readQuota != nil {
// 		u.ReadQuota = int64(*readQuota)
// 	}
// 	if writeQuota != nil {
// 		u.WriteQuota = int64(*writeQuota)
// 	}
// 	u.UpdatedOn = time.Now()
// 	_, err := db.DB.ModelContext(ctx, u).Column("read_quota", "write_quota", "updated_on").WherePK().Update()
// 	return u, err
// }

// RemoveUser removes a member from the organization
func (o *Organization) RemoveUser(ctx context.Context, userID uuid.UUID) error {
	// TODO: only if not last user (there's a check in resolver, but it should be part of db tx)

	// clear cache
	getUserOrganizationPermissionsCache().Clear(ctx, userID, o.OrganizationID)

	// delete
	return db.DB.WithContext(ctx).Delete(&PermissionsUsersOrganizations{
		UserID:         userID,
		OrganizationID: o.OrganizationID,
	})
}

// FindOrganizationPermissions retrieves all users' permissions for a given organization
func FindOrganizationPermissions(ctx context.Context, organizationID uuid.UUID) []*PermissionsUsersOrganizations {
	var permissions []*PermissionsUsersOrganizations
	err := db.DB.ModelContext(ctx, &permissions).Where("permissions_users_organizations.organization_id = ?", organizationID).Column("permissions_users_organizations.*", "User", "Organization").Select()
	if err != nil {
		panic(err)
	}

	return permissions
}
