package entity

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/go-pg/pg/v9"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/go-playground/validator.v9"

	"gitlab.com/beneath-hq/beneath/internal/hub"
	"gitlab.com/beneath-hq/beneath/pkg/log"
	"gitlab.com/beneath-hq/beneath/pkg/timeutil"
)

// Organization represents the entity that manages billing on behalf of its users
type Organization struct {
	OrganizationID    uuid.UUID  `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	Name              string     `sql:",unique,notnull",validate:"required,gte=3,lte=40"` // NOTE: when updating, clear stream cache
	DisplayName       string     `sql:",notnull",validate:"required,gte=1,lte=50"`
	Description       string     `validate:"omitempty,lte=255"`
	PhotoURL          string     `validate:"omitempty,url,lte=400"`
	UserID            *uuid.UUID `sql:",on_delete:restrict,type:uuid"`
	User              *User
	CreatedOn         time.Time `sql:",default:now()"`
	UpdatedOn         time.Time `sql:",default:now()"`
	PrepaidReadQuota  *int64    // bytes
	PrepaidWriteQuota *int64    // bytes
	ReadQuota         *int64    // bytes // NOTE: when updating value, clear secret cache
	WriteQuota        *int64    // bytes // NOTE: when updating value, clear secret cache
	Projects          []*Project
	Services          []*Service
	Users             []*User `pg:"fk:billing_organization_id"`

	// used to indicate requestor's permissions in resolvers
	Permissions *PermissionsUsersOrganizations `sql:"-"`
}

var (
	// used for validation
	organizationNameRegex     = regexp.MustCompile("^[_a-z][_a-z0-9]*$")
	organizationNameBlacklist = []string{
		"auth",
		"billing",
		"docs",
		"documentation",
		"explore",
		"health",
		"healthz",
		"instance",
		"instances",
		"organization",
		"organizations",
		"permissions",
		"project",
		"projects",
		"redirects",
		"secret",
		"secrets",
		"stream",
		"streams",
		"terminal",
		"user",
		"username",
		"users",
	}
)

func init() {
	// configure validation
	GetValidator().RegisterStructValidation(organizationValidation, Organization{})
}

// custom stream validation
func organizationValidation(sl validator.StructLevel) {
	s := sl.Current().Interface().(Organization)

	if !organizationNameRegex.MatchString(s.Name) {
		sl.ReportError(s.Name, "Name", "", "alphanumericorunderscore", "")
	}

	for _, blacklisted := range organizationNameBlacklist {
		if s.Name == blacklisted {
			sl.ReportError(s.Name, "Name", "", "blacklisted", "")
			break
		}
	}
}

// FindOrganization finds a organization by ID
func FindOrganization(ctx context.Context, organizationID uuid.UUID) *Organization {
	organization := &Organization{
		OrganizationID: organizationID,
	}
	err := hub.DB.ModelContext(ctx, organization).WherePK().
		Column(
			"organization.*",
			"User",
			"Services", // only necessary if has permissions
			"Projects",
		).Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return organization
}

// FindOrganizationByName finds a organization by name
func FindOrganizationByName(ctx context.Context, name string) *Organization {
	organization := &Organization{}
	err := hub.DB.ModelContext(ctx, organization).
		Where("lower(organization.name) = lower(?)", name).
		Column(
			"organization.*",
			"User",
			"User.BillingOrganization",
			"Services", // only necessary if has permissions
			"Projects",
		).Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return organization
}

// FindOrganizationByUserID finds an organization by its personal user ID (if set)
func FindOrganizationByUserID(ctx context.Context, userID uuid.UUID) *Organization {
	organization := &Organization{}
	err := hub.DB.ModelContext(ctx, organization).
		Where("organization.user_id = ?", userID).
		Column(
			"organization.*",
			"User",
			"User.BillingOrganization",
			"Services", // only necessary if has permissions
			"Projects",
		).Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return organization
}

// FindAllOrganizations returns all active organizations
func FindAllOrganizations(ctx context.Context) []*Organization {
	var organizations []*Organization
	err := hub.DB.ModelContext(ctx, &organizations).
		Column("organization.*").
		Select()
	if err != nil {
		panic(err)
	}
	return organizations
}

// IsOrganization implements gql.Organization
func (o *Organization) IsOrganization() {}

// IsMulti returns true if o is a multi-user organization
func (o *Organization) IsMulti() bool {
	return o.UserID == nil
}

// IsBillingOrganizationForUser returns true if o is also the billing org for the user it represents.
// It panics if called on a non-personal organization
func (o *Organization) IsBillingOrganizationForUser() bool {
	if o.UserID == nil {
		panic(fmt.Errorf("Called IsBillingOrganizationForUser on non-personal organization"))
	}
	return o.User.BillingOrganizationID == o.OrganizationID
}

// StripPrivateProjects removes private projects from o.Projects (no changes in database, just the loaded object)
func (o *Organization) StripPrivateProjects() {
	for i, p := range o.Projects {
		if !p.Public {
			n := len(o.Projects)
			o.Projects[n-1], o.Projects[i] = o.Projects[i], o.Projects[n-1]
			o.Projects = o.Projects[:n-1]
		}
	}
}

// CreateWithUser creates an organization and makes user a member
func (o *Organization) CreateWithUser(ctx context.Context, userID uuid.UUID, view bool, create bool, admin bool) error {
	defaultBillingPlan := FindDefaultBillingPlan(ctx)

	o.PrepaidReadQuota = &defaultBillingPlan.ReadQuota
	o.PrepaidWriteQuota = &defaultBillingPlan.WriteQuota
	o.ReadQuota = &defaultBillingPlan.ReadQuota
	o.WriteQuota = &defaultBillingPlan.WriteQuota

	// validate
	err := GetValidator().Struct(o)
	if err != nil {
		return err
	}

	// create organization and PermissionsUsersOrganizations in one transaction
	err = hub.DB.WithContext(ctx).RunInTransaction(func(tx *pg.Tx) error {
		// insert org
		_, err := tx.Model(o).Insert()
		if err != nil {
			return err
		}

		// connect org to userID
		err = tx.Insert(&PermissionsUsersOrganizations{
			UserID:         userID,
			OrganizationID: o.OrganizationID,
			View:           view,
			Create:         create,
			Admin:          admin,
		})
		if err != nil {
			return err
		}

		// create billing info
		bi := &BillingInfo{
			OrganizationID: o.OrganizationID,
			BillingPlanID:  defaultBillingPlan.BillingPlanID,
		}
		_, err = tx.Model(bi).Insert()
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	// log new organization
	log.S.Infow(
		"control created non-personal organization",
		"organization_id", o.OrganizationID,
	)

	return nil
}

// UpdateDetails updates the organization's name, display name, description and/or photo
func (o *Organization) UpdateDetails(ctx context.Context, name *string, displayName *string, description *string, photoURL *string) error {
	if name != nil {
		o.Name = *name
	}
	if displayName != nil {
		o.DisplayName = *displayName
	}
	if description != nil {
		o.Description = *description
	}
	if photoURL != nil {
		o.PhotoURL = *photoURL
	}

	// validate
	err := GetValidator().Struct(o)
	if err != nil {
		return err
	}

	o.UpdatedOn = time.Now()
	_, err = hub.DB.ModelContext(ctx, o).
		Column(
			"name",
			"display_name",
			"description",
			"photo_url",
			"updated_on",
		).WherePK().Update()
	if err != nil {
		return err
	}

	// flush all cached streams if name changed
	if name != nil {
		getStreamCache().ClearForOrganization(ctx, o.OrganizationID)
	}

	return nil
}

// UpdateQuotas updates the quotas enforced upon the organization
func (o *Organization) UpdateQuotas(ctx context.Context, readQuota *int64, writeQuota *int64) error {
	// set fields
	o.ReadQuota = readQuota
	o.WriteQuota = writeQuota

	// validate
	err := GetValidator().Struct(o)
	if err != nil {
		return err
	}

	// update
	o.UpdatedOn = time.Now()
	_, err = hub.DB.ModelContext(ctx, o).Column("read_quota", "write_quota", "updated_on").WherePK().Update()

	// clear cache for the organization's users' secrets
	getSecretCache().ClearForOrganization(ctx, o.OrganizationID)

	return err
}

// UpdatePrepaidQuotas updates the organization's prepaid quotas
func (o *Organization) UpdatePrepaidQuotas(ctx context.Context, billingPlan *BillingPlan) error {
	numSeats := int64(len(o.Users))
	prepaidReadQuota := billingPlan.BaseReadQuota + billingPlan.SeatReadQuota*numSeats
	prepaidWriteQuota := billingPlan.BaseWriteQuota + billingPlan.SeatWriteQuota*numSeats

	// set fields
	o.PrepaidReadQuota = &prepaidReadQuota
	o.PrepaidWriteQuota = &prepaidWriteQuota
	o.UpdatedOn = time.Now()

	// update
	_, err := hub.DB.WithContext(ctx).Model(o).
		Column("prepaid_read_quota", "prepaid_write_quota", "updated_on").
		WherePK().
		Update()
	if err != nil {
		return err
	}

	return err
}

// Delete deletes the organization
// this will fail if there are any users, services, or projects that are still tied to the organization
func (o *Organization) Delete(ctx context.Context) error {
	// NOTE: effectively, this resolver doesn't work, as there will always be one admin member
	// TODO: check it's empty and has just one admin user, trigger bill, make the admin leave, delete org

	err := hub.DB.WithContext(ctx).Delete(o)
	if err != nil {
		return err
	}

	return nil
}

// TransferUser transfers a user in o to another organization
func (o *Organization) TransferUser(ctx context.Context, user *User, targetOrg *Organization) error {
	// can be from personal to multi, multi to personal, multi to multi

	// Ensure the last admin member isn't leaving a multi-user org
	if o.IsMulti() {
		foundRemainingAdmin := false
		members, err := FindOrganizationMembers(ctx, o.OrganizationID)
		if err != nil {
			return err
		}
		for _, member := range members {
			if member.Admin && member.UserID != user.UserID {
				foundRemainingAdmin = true
				break
			}
		}
		if !foundRemainingAdmin {
			return fmt.Errorf("Cannot transfer user because it would leave the organization without an admin")
		}
	}

	// update user
	user.BillingOrganization = targetOrg
	user.BillingOrganizationID = targetOrg.OrganizationID
	user.ReadQuota = nil
	user.WriteQuota = nil
	user.UpdatedOn = time.Now()
	_, err := hub.DB.WithContext(ctx).Model(user).
		Column("billing_organization_id", "read_quota", "write_quota", "updated_on").
		WherePK().
		Update()
	if err != nil {
		return err
	}
	getSecretCache().ClearForUser(ctx, user.UserID)

	// find target billing info
	targetBillingInfo := FindBillingInfo(ctx, targetOrg.OrganizationID)
	if targetBillingInfo == nil {
		return fmt.Errorf("Couldn't find billing info for target organization")
	}

	// add prorated seat to the target organization's next month's bill
	billingTime := timeutil.Next(time.Now(), targetBillingInfo.BillingPlan.Period)
	err = commitProratedSeatsToBill(ctx, targetBillingInfo, billingTime, []*User{user})
	if err != nil {
		panic("unable to commit prorated seat to bill")
	}

	// increment the target organization's prepaid quota by the seat quota
	// we do this now because we want to show the new usage capacity in the UI as soon as possible
	if targetBillingInfo.BillingPlan.SeatReadQuota > 0 || targetBillingInfo.BillingPlan.SeatWriteQuota > 0 {
		newPrepaidReadQuota := *targetOrg.PrepaidReadQuota + targetBillingInfo.BillingPlan.SeatReadQuota
		newPrepaidWriteQuota := *targetOrg.PrepaidWriteQuota + targetBillingInfo.BillingPlan.SeatWriteQuota
		targetOrg.PrepaidReadQuota = &newPrepaidReadQuota
		targetOrg.PrepaidWriteQuota = &newPrepaidWriteQuota
		targetOrg.UpdatedOn = time.Now()

		_, err = hub.DB.WithContext(ctx).Model(targetOrg).
			Column("prepaid_read_quota", "prepaid_write_quota", "updated_on").
			WherePK().
			Update()
		if err != nil {
			return err
		}
	}

	return nil
}

// TransferProject transfers a project in o to another organization
func (o *Organization) TransferProject(ctx context.Context, project *Project, targetOrg *Organization) error {
	project.Organization = targetOrg
	project.OrganizationID = targetOrg.OrganizationID
	project.UpdatedOn = time.Now()
	_, err := hub.DB.ModelContext(ctx, project).Column("organization_id", "updated_on").WherePK().Update()
	if err != nil {
		return err
	}

	// stream cache includes organization name, so invalidate cache
	getStreamCache().ClearForProject(ctx, project.ProjectID)

	return nil
}

// TransferService transfers a service in o to another organization
func (o *Organization) TransferService(ctx context.Context, service *Service, targetOrg *Organization) error {
	// update organization
	service.Organization = targetOrg
	service.OrganizationID = targetOrg.OrganizationID
	service.UpdatedOn = time.Now()
	_, err := hub.DB.ModelContext(ctx, service).Column("organization_id", "updated_on").WherePK().Update()
	if err != nil {
		return err
	}

	// secret cache includes organization's quotas
	getSecretCache().ClearForService(ctx, service.ServiceID)

	return nil
}
