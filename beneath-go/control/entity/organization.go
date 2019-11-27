package entity

import (
	"context"
	"regexp"
	"time"

	"github.com/go-pg/pg/v9"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/go-playground/validator.v9"

	"github.com/beneath-core/beneath-go/db"
)

// Organization represents the entity that manages billing on behalf of its users
type Organization struct {
	OrganizationID   uuid.UUID `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	Name             string    `sql:",unique,notnull",validate:"required,gte=1,lte=40"`
	Personal         bool      `sql:",notnull"`
	StripeCustomerID string
	BillingPlanID    uuid.UUID `sql:"on_delete:RESTRICT,notnull,type:uuid"`
	BillingPlan      *BillingPlan
	PaymentMethod    PaymentMethod
	CreatedOn        time.Time `sql:",default:now()"`
	UpdatedOn        time.Time `sql:",default:now()"`
	DeletedOn        time.Time
	Services         []*Service
	Users            []*User `pg:"many2many:permissions_users_organizations,fk:organization_id,joinFK:user_id"`
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
	err := db.DB.ModelContext(ctx, organization).WherePK().Column("organization.*", "Services", "Users", "BillingPlan").Select()
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
		Column("organization.*", "Services", "Users", "BillingPlan").
		Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return organization
}

// FindAllOrganizations returns all organizations
func FindAllOrganizations(ctx context.Context) []Organization {
	var organizations []Organization
	err := db.DB.ModelContext(ctx, &organizations).Column("organization.*").Select()
	if err != nil {
		panic(err)
	}
	return organizations
}

// CreateOrganizationWithUser creates an organization
func CreateOrganizationWithUser(ctx context.Context, name string, userID uuid.UUID) (*Organization, error) {
	// create
	org := &Organization{
		Name: name,
	}

	// validate
	err := GetValidator().Struct(org)
	if err != nil {
		return nil, err
	}

	// create organization
	err = db.DB.WithContext(ctx).RunInTransaction(func(tx *pg.Tx) error {
		// insert org
		_, err := tx.Model(org).Insert()
		if err != nil {
			return err
		}

		// add user
		err = tx.Insert(&PermissionsUsersOrganizations{
			UserID:         userID,
			OrganizationID: org.OrganizationID,
			View:           true,
			Admin:          true,
		})
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return org, nil
}

// AddUser makes user a member of organization
func (o *Organization) AddUser(ctx context.Context, userID uuid.UUID, view bool, admin bool) error {
	return db.DB.WithContext(ctx).Insert(&PermissionsUsersOrganizations{
		UserID:         userID,
		OrganizationID: o.OrganizationID,
		View:           view,
		Admin:          admin,
	})
}

// ChangeUserPermissions changes a user's permissions within the organization
func (o *Organization) ChangeUserPermissions(ctx context.Context, userID uuid.UUID, view bool, admin bool) error {
	// TODO: convert permissions.Update (below) to this function
	return nil
}

// Update changes a user's permissions within the organization
// func (p *PermissionsUsersOrganizations) Update(ctx context.Context, view *bool, admin *bool) (*PermissionsUsersOrganizations, error) {
// 	// Q: this won't fuck up the object "p" that I have, right?
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

// UpdateBillingPlanID updates an organization's billing plan ID
func (o *Organization) UpdateBillingPlanID(ctx context.Context, billingPlanID uuid.UUID) (*Organization, error) {
	o.BillingPlanID = billingPlanID
	o.UpdatedOn = time.Now()
	_, err := db.DB.ModelContext(ctx, o).Column("billing_plan_id", "updated_on").WherePK().Update()
	return o, err
}

// UpdateStripeCustomerID updates an organization's Stripe Customer ID
func (o *Organization) UpdateStripeCustomerID(ctx context.Context, stripeCustomerID string) error {
	o.StripeCustomerID = stripeCustomerID
	o.UpdatedOn = time.Now()
	_, err := db.DB.ModelContext(ctx, o).Column("stripe_customer_id", "updated_on").WherePK().Update()
	return err
}

// UpdatePaymentMethod updates an organization's payment method type
func (o *Organization) UpdatePaymentMethod(ctx context.Context, paymentMethod PaymentMethod) error {
	o.PaymentMethod = paymentMethod
	o.UpdatedOn = time.Now()
	_, err := db.DB.ModelContext(ctx, o).Column("payment_method", "updated_on").WherePK().Update()
	return err
}
