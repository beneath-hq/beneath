package entity

import (
	"context"

	"github.com/go-pg/pg/v9/orm"
	uuid "github.com/satori/go.uuid"

	"gitlab.com/beneath-hq/beneath/hub"
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
	Create         bool `sql:",notnull"`
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

// FindPermissionsUsersProjects finds a user's permissions for a project
func FindPermissionsUsersProjects(ctx context.Context, userID uuid.UUID, projectID uuid.UUID) *PermissionsUsersProjects {
	permissions := &PermissionsUsersProjects{
		UserID:    userID,
		ProjectID: projectID,
	}
	err := hub.DB.ModelContext(ctx, permissions).
		WherePK().
		Select()

	if !AssertFoundOne(err) {
		return nil
	}

	return permissions
}

// FindPermissionsUsersOrganizations finds a user's permissions for an organization
func FindPermissionsUsersOrganizations(ctx context.Context, userID uuid.UUID, organizationID uuid.UUID) *PermissionsUsersOrganizations {
	permissions := &PermissionsUsersOrganizations{
		UserID:         userID,
		OrganizationID: organizationID,
	}
	err := hub.DB.ModelContext(ctx, permissions).
		WherePK().
		Select()

	if !AssertFoundOne(err) {
		return nil
	}

	return permissions
}

// FindPermissionsServicesStreams finds a service's permissions for a stream
func FindPermissionsServicesStreams(ctx context.Context, serviceID uuid.UUID, streamID uuid.UUID) *PermissionsServicesStreams {
	permissions := &PermissionsServicesStreams{
		ServiceID: serviceID,
		StreamID:  streamID,
	}
	err := hub.DB.ModelContext(ctx, permissions).
		WherePK().
		Select()

	if !AssertFoundOne(err) {
		return nil
	}

	return permissions
}

// Update upserts permissions (or deletes them if all are falsy)
func (p *PermissionsUsersProjects) Update(ctx context.Context, view *bool, create *bool, admin *bool) error {
	if view != nil {
		p.View = *view
	}
	if create != nil {
		p.Create = *create
	}
	if admin != nil {
		p.Admin = *admin
	}

	// if all are falsy, delete the permiission (if it exists), else update
	if !p.View && !p.Create && !p.Admin {
		_, err := hub.DB.ModelContext(ctx, p).WherePK().Delete()
		if err != nil {
			return err
		}
	} else {
		// build upsert
		q := hub.DB.ModelContext(ctx, p).OnConflict("(user_id, project_id) DO UPDATE")
		if view != nil {
			q = q.Set("view = EXCLUDED.view")
		}
		if create != nil {
			q = q.Set(`"create" = EXCLUDED."create"`)
		}
		if admin != nil {
			q = q.Set("admin = EXCLUDED.admin")
		}

		// run upsert
		_, err := q.Insert()
		if err != nil {
			return err
		}
	}

	// clear cache
	getUserProjectPermissionsCache().Clear(ctx, p.UserID, p.ProjectID)

	return nil
}

// Update upserts permissions (or deletes them if all are falsy)
func (p *PermissionsUsersOrganizations) Update(ctx context.Context, view *bool, create *bool, admin *bool) error {
	if view != nil {
		p.View = *view
	}
	if create != nil {
		p.Create = *create
	}
	if admin != nil {
		p.Admin = *admin
	}

	// if all are falsy, delete the permiission (if it exists), else update
	if !p.View && !p.Create && !p.Admin {
		_, err := hub.DB.ModelContext(ctx, p).WherePK().Delete()
		if err != nil {
			return err
		}
	} else {
		// build upsert
		q := hub.DB.ModelContext(ctx, p).OnConflict("(user_id, organization_id) DO UPDATE")
		if view != nil {
			q = q.Set("view = EXCLUDED.view")
		}
		if create != nil {
			q = q.Set(`"create" = EXCLUDED."create"`)
		}
		if admin != nil {
			q = q.Set("admin = EXCLUDED.admin")
		}

		// run upsert
		_, err := q.Insert()
		if err != nil {
			return err
		}
	}

	// clear cache
	getUserOrganizationPermissionsCache().Clear(ctx, p.UserID, p.OrganizationID)

	return nil
}

// Update upserts permissions (or deletes them if all are falsy)
func (p *PermissionsServicesStreams) Update(ctx context.Context, read *bool, write *bool) error {
	if read != nil {
		p.Read = *read
	}
	if write != nil {
		p.Write = *write
	}

	// if all are falsy, delete the permiission (if it exists), else update
	if !p.Read && !p.Write {
		_, err := hub.DB.ModelContext(ctx, p).WherePK().Delete()
		if err != nil {
			return err
		}
	} else {
		// build upsert
		q := hub.DB.ModelContext(ctx, p).OnConflict("(service_id, stream_id) DO UPDATE")
		if read != nil {
			q = q.Set("read = EXCLUDED.read")
		}
		if write != nil {
			q = q.Set("write = EXCLUDED.write")
		}

		// run upsert
		_, err := q.Insert()
		if err != nil {
			return err
		}
	}

	// clear cache
	getServiceStreamPermissionsCache().Clear(ctx, p.ServiceID, p.StreamID)

	return nil
}
