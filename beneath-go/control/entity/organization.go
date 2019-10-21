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
	OrganizationID uuid.UUID `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	Name           string    `sql:",unique,notnull",validate:"required,gte=1,lte=40"`
	CreatedOn      time.Time `sql:",default:now()"`
	UpdatedOn      time.Time `sql:",default:now()"`
	DeletedOn      time.Time
	Services       []*Service
	Users          []*User `pg:"many2many:permissions_users_organizations,fk:organization_id,joinFK:user_id"`
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

// RemoveUser removes a member from the organization
func (o *Organization) RemoveUser(ctx context.Context, userID uuid.UUID) error {
	// TODO remove from cache?
	// TODO only if not last user (there's a check in resolver, but it should be part of db tx)?
	return db.DB.WithContext(ctx).Delete(&PermissionsUsersOrganizations{
		UserID:         userID,
		OrganizationID: o.OrganizationID,
	})
}
