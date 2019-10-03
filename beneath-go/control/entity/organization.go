package entity

import (
	"context"
	"time"

	"github.com/beneath-core/beneath-go/db"
	"github.com/go-pg/pg/v9"
	uuid "github.com/satori/go.uuid"
)

// Organization represents the entity that manages billing on behalf of its users
type Organization struct {
	OrganizationID uuid.UUID `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	Name           string    `sql:",notnull",validate:"required,gte=1,lte=40"`
	CreatedOn      time.Time `sql:",default:now()"`
	UpdatedOn      time.Time `sql:",default:now()"`
	DeletedOn      time.Time
	Services       []*Service
	Users          []*User `pg:"many2many:permissions_users_organizations,fk:organization_id,joinFK:user_id"`
}

// Create creates an organization
func (o *Organization) Create(ctx context.Context, name string) error {
	// validate
	err := GetValidator().Struct(o) // check this out
	if err != nil {
		return err
	}

	// create organization
	return db.DB.WithContext(ctx).RunInTransaction(func(tx *pg.Tx) error {
		// insert project
		_, err := tx.Model(o).Insert()
		if err != nil {
			return err
		}

		// done
		return nil
	})
}

// FindOrganizationByName finds a organization by name
func FindOrganizationByName(ctx context.Context, name string) *Organization {
	organization := &Organization{}
	err := db.DB.ModelContext(ctx, organization).
		Where("lower(organization.name) = lower(?)", name).
		Column("organization.*", "Services", "Users").
		Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return organization
}