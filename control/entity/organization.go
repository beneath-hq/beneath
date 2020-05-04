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
	OrganizationID uuid.UUID `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	Name           string    `sql:",unique,notnull",validate:"required,gte=1,lte=40"`
	Personal       bool      `sql:",notnull"`
	CreatedOn      time.Time `sql:",default:now()"`
	UpdatedOn      time.Time `sql:",default:now()"`
	ReadQuota      *int64
	WriteQuota     *int64
	Projects       []*Project
	Services       []*Service
	Users          []*User `pg:"many2many:permissions_users_organizations,fk:user_id,joinFK:organization_id"`
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
	err := hub.DB.ModelContext(ctx, organization).WherePK().Column("organization.*", "Projects", "Services", "Users").Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return organization
}

// FindOrganizationByName finds a organization by name
func FindOrganizationByName(ctx context.Context, name string) *Organization {
	organization := &Organization{}
	err := hub.DB.ModelContext(ctx, organization).
		Where("lower(name) = lower(?)", name).
		Column("organization.*", "Projects", "Services", "Users"). // Q: unable to find Users
		Select()
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

// FindOrganizationPermissions retrieves all users' permissions for a given organization
func FindOrganizationPermissions(ctx context.Context, organizationID uuid.UUID) []*PermissionsUsersOrganizations {
	var permissions []*PermissionsUsersOrganizations
	err := hub.DB.ModelContext(ctx, &permissions).Where("permissions_users_organizations.organization_id = ?", organizationID).Column("permissions_users_organizations.*", "User", "Organization").Select()
	if err != nil {
		panic(err)
	}

	return permissions
}

// CreateWithUser creates an organization and makes user a member
func (o *Organization) CreateWithUser(ctx context.Context, userID uuid.UUID, view bool, create bool, admin bool) error {
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

		// get default billing plan
		defaultBillingPlan := FindDefaultBillingPlan(ctx)

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
		"control created enterprise organization",
		"organization_id", o.OrganizationID,
	)

	return nil
}

// UpdateName updates an organization's name
func (o *Organization) UpdateName(ctx context.Context, name string) (*Organization, error) {
	o.Name = name
	o.UpdatedOn = time.Now()

	_, err := hub.DB.ModelContext(ctx, o).
		Column("name", "updated_on").
		WherePK().
		Update()
	if err != nil {
		return nil, err
	}

	return o, nil
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
	if !o.Personal {
		foundRemainingAdmin := false
		members := FindOrganizationPermissions(ctx, o.OrganizationID)
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

	// find new billing info
	targetBillingInfo := FindBillingInfo(ctx, targetOrg.OrganizationID)
	if targetBillingInfo == nil {
		return fmt.Errorf("Couldn't find billing info for target organization")
	}

	// commit user's current usage to the old organization's bill
	err := commitCurrentUsageToNextBill(ctx, o.OrganizationID, UserEntityKind, user.UserID, user.Username, false)
	if err != nil {
		return err
	}

	// update user
	user.BillingOrganization = targetOrg
	user.BillingOrganizationID = targetOrg.OrganizationID
	user.ReadQuota = &targetBillingInfo.BillingPlan.SeatReadQuota
	user.WriteQuota = &targetBillingInfo.BillingPlan.SeatWriteQuota
	user.UpdatedOn = time.Now()
	_, err = hub.DB.WithContext(ctx).Model(user).
		Column("billing_organization_id", "read_quota", "write_quota", "updated_on").
		WherePK().
		Update()
	if err != nil {
		return err
	}

	// commit usage credit to the new organization's bill for the user's current month's usage
	err = commitCurrentUsageToNextBill(ctx, targetOrg.OrganizationID, UserEntityKind, user.UserID, user.Username, true)
	if err != nil {
		panic("unable to commit usage credit to bill")
	}

	// add prorated seat to the new organization's next month's bill
	billingTime := timeutil.Next(time.Now(), targetBillingInfo.BillingPlan.Period)
	err = commitProratedSeatsToBill(ctx, targetOrg.OrganizationID, billingTime, targetBillingInfo.BillingPlan, []uuid.UUID{user.UserID}, []string{user.Username}, false)
	if err != nil {
		panic("unable to commit prorated seat to bill")
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

	return nil
}

// TransferService transfers a service in o to another organization
func (o *Organization) TransferService(ctx context.Context, service *Service, targetOrg *Organization) error {
	// commit current usage to the old organization's bill
	err := commitCurrentUsageToNextBill(ctx, o.OrganizationID, ServiceEntityKind, service.ServiceID, service.Name, false)
	if err != nil {
		return err
	}

	// update organization
	service.Organization = targetOrg
	service.OrganizationID = targetOrg.OrganizationID
	service.UpdatedOn = time.Now()
	_, err = hub.DB.ModelContext(ctx, service).Column("organization_id", "updated_on").WherePK().Update()
	if err != nil {
		return err
	}

	// commit usage credit to the new organization's bill for the service's current month's usage
	err = commitCurrentUsageToNextBill(ctx, targetOrg.OrganizationID, ServiceEntityKind, service.ServiceID, service.Name, true)
	if err != nil {
		return err
	}

	return nil
}
