package entity

import (
	"context"

	"gitlab.com/beneath-hq/beneath/internal/hub"
	"github.com/go-pg/pg/v9/orm"
	uuid "github.com/satori/go.uuid"
)

// PermissionsUsersProjects represents the many-to-many relationship between users and projects
type PermissionsUsersProjects struct {
	tableName struct{}  `sql:"permissions_users_projects"`
	UserID    uuid.UUID `sql:"on_delete:CASCADE,pk,type:uuid"`
	User      *User
	ProjectID uuid.UUID `sql:"on_delete:CASCADE,pk,type:uuid"`
	Project   *Project
	View      bool `sql:",notnull"`
	Create    bool `sql:",notnull"`
	Admin     bool `sql:",notnull"`
}

// PermissionsUsersOrganizations represents the many-to-one relationship between users and organizations
type PermissionsUsersOrganizations struct {
	tableName      struct{}  `sql:"permissions_users_organizations"`
	UserID         uuid.UUID `sql:"on_delete:CASCADE,pk,type:uuid"`
	User           *User
	OrganizationID uuid.UUID `sql:"on_delete:CASCADE,pk,type:uuid"`
	Organization   *Organization
	View           bool `sql:",notnull"`
	Admin          bool `sql:",notnull"`
}

// PermissionsServicesStreams represnts the many-to-many relationship between services and projects
type PermissionsServicesStreams struct {
	tableName struct{}  `sql:"permissions_services_streams"`
	ServiceID uuid.UUID `sql:"on_delete:CASCADE,pk,type:uuid"`
	Service   *Service
	StreamID  uuid.UUID `sql:"on_delete:CASCADE,pk,type:uuid"`
	Stream    *Stream
	Read      bool `sql:",notnull"`
	Write     bool `sql:",notnull"`
}

func init() {
	orm.RegisterTable((*PermissionsUsersProjects)(nil))
	orm.RegisterTable((*PermissionsUsersOrganizations)(nil))
	orm.RegisterTable((*PermissionsServicesStreams)(nil))
}

// FindPermissionsUsersOrganizations finds a organization by ID
func FindPermissionsUsersOrganizations(ctx context.Context, userID uuid.UUID, organizationID uuid.UUID) *PermissionsUsersOrganizations {
	permissions := &PermissionsUsersOrganizations{
		UserID:         userID,
		OrganizationID: organizationID,
	}
	err := hub.DB.ModelContext(ctx, permissions).
		WherePK().
		Column("permissions_users_organizations.*", "User", "Organization").
		Select()

	if !AssertFoundOne(err) {
		return nil
	}

	return permissions
}
