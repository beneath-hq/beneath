package permissions

import (
	"context"

	uuid "github.com/satori/go.uuid"

	"github.com/go-redis/redis/v7"
	"gitlab.com/beneath-hq/beneath/bus"
	"gitlab.com/beneath-hq/beneath/infrastructure/db"
	"gitlab.com/beneath-hq/beneath/models"
)

// Service provides functionality for checking users' and services' permissions for resources
type Service struct {
	Bus   *bus.Bus
	DB    db.DB
	Redis *redis.Client

	userOrganizationCache *Cache
	userProjectCache      *Cache
	serviceStreamCache    *Cache
}

// New creates a new permissions service
func New(bus *bus.Bus, db db.DB, redis *redis.Client) *Service {
	s := &Service{
		Bus:   bus,
		DB:    db,
		Redis: redis,
	}
	s.initCaches()
	return s
}

// FindPermissionsUsersProjects finds a user's permissions for a project
func (s *Service) FindPermissionsUsersProjects(ctx context.Context, userID uuid.UUID, projectID uuid.UUID) *models.PermissionsUsersProjects {
	permissions := &models.PermissionsUsersProjects{
		UserID:    userID,
		ProjectID: projectID,
	}
	err := s.DB.GetDB(ctx).ModelContext(ctx, permissions).
		WherePK().
		Select()
	if !db.AssertFoundOne(err) {
		return nil
	}
	return permissions
}

// FindPermissionsUsersOrganizations finds a user's permissions for an organization
func (s *Service) FindPermissionsUsersOrganizations(ctx context.Context, userID uuid.UUID, organizationID uuid.UUID) *models.PermissionsUsersOrganizations {
	permissions := &models.PermissionsUsersOrganizations{
		UserID:         userID,
		OrganizationID: organizationID,
	}
	err := s.DB.GetDB(ctx).ModelContext(ctx, permissions).
		WherePK().
		Select()
	if !db.AssertFoundOne(err) {
		return nil
	}
	return permissions
}

// FindPermissionsServicesStreams finds a service's permissions for a stream
func (s *Service) FindPermissionsServicesStreams(ctx context.Context, serviceID uuid.UUID, streamID uuid.UUID) *models.PermissionsServicesStreams {
	permissions := &models.PermissionsServicesStreams{
		ServiceID: serviceID,
		StreamID:  streamID,
	}
	err := s.DB.GetDB(ctx).ModelContext(ctx, permissions).
		WherePK().
		Select()
	if !db.AssertFoundOne(err) {
		return nil
	}
	return permissions
}

// UpdateUserProjectPermission upserts permissions (or deletes them if all are falsy)
func (s *Service) UpdateUserProjectPermission(ctx context.Context, p *models.PermissionsUsersProjects, view *bool, create *bool, admin *bool) error {
	if view != nil {
		p.View = *view
	}
	if create != nil {
		p.Create = *create
	}
	if admin != nil {
		p.Admin = *admin
	}

	// if all are falsy, delete the permission (if it exists), else update
	if !p.View && !p.Create && !p.Admin {
		_, err := s.DB.GetDB(ctx).ModelContext(ctx, p).WherePK().Delete()
		if err != nil {
			return err
		}
	} else {
		// build upsert
		q := s.DB.GetDB(ctx).ModelContext(ctx, p).OnConflict("(user_id, project_id) DO UPDATE")
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
	s.userProjectCache.Clear(ctx, p.UserID, p.ProjectID)

	return nil
}

// UpdateUserOrganizationPermission upserts permissions (or deletes them if all are falsy)
func (s *Service) UpdateUserOrganizationPermission(ctx context.Context, p *models.PermissionsUsersOrganizations, view *bool, create *bool, admin *bool) error {
	if view != nil {
		p.View = *view
	}
	if create != nil {
		p.Create = *create
	}
	if admin != nil {
		p.Admin = *admin
	}

	// if all are falsy, delete the permission (if it exists), else update
	if !p.View && !p.Create && !p.Admin {
		_, err := s.DB.GetDB(ctx).ModelContext(ctx, p).WherePK().Delete()
		if err != nil {
			return err
		}
	} else {
		// build upsert
		q := s.DB.GetDB(ctx).ModelContext(ctx, p).OnConflict("(user_id, organization_id) DO UPDATE")
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
	s.userOrganizationCache.Clear(ctx, p.UserID, p.OrganizationID)

	return nil
}

// UpdateServiceStreamPermission upserts permissions (or deletes them if all are falsy)
func (s *Service) UpdateServiceStreamPermission(ctx context.Context, p *models.PermissionsServicesStreams, read *bool, write *bool) error {
	if read != nil {
		p.Read = *read
	}
	if write != nil {
		p.Write = *write
	}

	// if all are falsy, delete the permission (if it exists), else update
	if !p.Read && !p.Write {
		_, err := s.DB.GetDB(ctx).ModelContext(ctx, p).WherePK().Delete()
		if err != nil {
			return err
		}
	} else {
		// build upsert
		q := s.DB.GetDB(ctx).ModelContext(ctx, p).OnConflict("(service_id, stream_id) DO UPDATE")
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
	s.serviceStreamCache.Clear(ctx, p.ServiceID, p.StreamID)

	return nil
}
