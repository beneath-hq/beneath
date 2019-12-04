package entity

import (
	"context"

	"github.com/beneath-core/beneath-go/db"
	"github.com/go-pg/pg/v9/orm"
	uuid "github.com/satori/go.uuid"
)

// PermissionsUsersProjects represnts the many-to-many relationship between users and projects
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

// PermissionsUsersOrganizations represnts the many-to-many relationship between users and organizations
type PermissionsUsersOrganizations struct {
	tableName      struct{}  `sql:"permissions_users_organizations"`
	UserID         uuid.UUID `sql:"on_delete:CASCADE,pk,type:uuid"`
	User           *User
	OrganizationID uuid.UUID `sql:"on_delete:CASCADE,pk,type:uuid"`
	Organization   *Organization
	View           bool `sql:",notnull"`
	Admin          bool `sql:",notnull"`
}

// PermissionsServicesStreams represnts the many-to-many relationship between users and projects
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

// TODO(review): Remove this -- we have CachedUserOrganizationPermissions. Also, I think it would be possible to not need this in the first place.

// FindPermissionsUsersOrganizations finds permissions by userID and organizationID
func FindPermissionsUsersOrganizations(ctx context.Context, userID uuid.UUID, organizationID uuid.UUID) *PermissionsUsersOrganizations {
	permissions := &PermissionsUsersOrganizations{
		UserID:         userID,
		OrganizationID: organizationID,
	}
	err := db.DB.ModelContext(ctx, permissions).WherePK().Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return permissions
}
