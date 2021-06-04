package service

import (
	"context"
	"time"

	"github.com/beneath-hq/beneath/bus"
	"github.com/beneath-hq/beneath/infra/db"
	"github.com/beneath-hq/beneath/models"
	uuid "github.com/satori/go.uuid"
)

// Service contains functionality for finding and creating services.
// YES, THIS IS CONFUSING! In the codebase, a "service" wraps functionality
// for a specific domain of the app. Meanwhile in Beneath, a "service" is a
// non-user account, also known as a "service account" in eg. GCP.
type Service struct {
	Bus *bus.Bus
	DB  db.DB
}

// New creates a new user service
func New(bus *bus.Bus, db db.DB) *Service {
	return &Service{
		Bus: bus,
		DB:  db,
	}
}

// FindService returns the matching service or nil
func (s *Service) FindService(ctx context.Context, serviceID uuid.UUID) *models.Service {
	service := &models.Service{
		ServiceID: serviceID,
	}
	err := s.DB.GetDB(ctx).ModelContext(ctx, service).
		Column(
			"service.*",
			"Project",
			"Project.Organization",
		).
		WherePK().
		Select()
	if !db.AssertFoundOne(err) {
		return nil
	}
	return service
}

// FindServiceByOrganizationProjectAndName returns the matching service or nil
func (s *Service) FindServiceByOrganizationProjectAndName(ctx context.Context, organizationName, projectName, serviceName string) *models.Service {
	service := &models.Service{}
	err := s.DB.GetDB(ctx).ModelContext(ctx, service).
		Column(
			"service.*",
			"Project",
			"Project.Organization",
		).
		Where("lower(project__organization.name) = lower(?)", organizationName).
		Where("lower(project.name) = lower(?)", projectName).
		Where("lower(service.name) = lower(?)", serviceName).
		Select()
	if !db.AssertFoundOne(err) {
		return nil
	}
	return service
}

// Create creates the service
func (s *Service) Create(ctx context.Context, service *models.Service, description *string, sourceURL *string, readQuota *int64, writeQuota *int64, scanQuota *int64) error {
	// assign
	if description != nil {
		service.Description = *description
	}
	if sourceURL != nil {
		service.SourceURL = *sourceURL
	}
	if readQuota != nil {
		if *readQuota == 0 {
			service.ReadQuota = nil
		} else {
			service.ReadQuota = readQuota
		}
	}
	if writeQuota != nil {
		if *writeQuota == 0 {
			service.WriteQuota = nil
		} else {
			service.WriteQuota = writeQuota
		}
	}
	if scanQuota != nil {
		if *scanQuota == 0 {
			service.ScanQuota = nil
		} else {
			service.ScanQuota = scanQuota
		}
	}

	// validate
	err := service.Validate()
	if err != nil {
		return err
	}

	// insert
	_, err = s.DB.GetDB(ctx).ModelContext(ctx, service).Insert()
	if err != nil {
		return err
	}

	return nil
}

// Update updates the service info
func (s *Service) Update(ctx context.Context, service *models.Service, description *string, sourceURL *string, readQuota *int64, writeQuota *int64, scanQuota *int64) error {
	// assign
	if description != nil {
		service.Description = *description
	}
	if sourceURL != nil {
		service.SourceURL = *sourceURL
	}
	if readQuota != nil {
		if *readQuota == 0 {
			service.ReadQuota = nil
		} else {
			service.ReadQuota = readQuota
		}
	}
	if writeQuota != nil {
		if *writeQuota == 0 {
			service.WriteQuota = nil
		} else {
			service.WriteQuota = writeQuota
		}
	}
	if scanQuota != nil {
		if *scanQuota == 0 {
			service.ScanQuota = nil
		} else {
			service.ScanQuota = scanQuota
		}
	}

	// validate
	err := service.Validate()
	if err != nil {
		return err
	}

	// update
	service.UpdatedOn = time.Now()
	_, err = s.DB.GetDB(ctx).ModelContext(ctx, service).WherePK().Update()
	if err != nil {
		return err
	}

	// publish update event
	err = s.Bus.Publish(ctx, &models.ServiceUpdatedEvent{
		Service: service,
	})
	if err != nil {
		return err
	}

	return nil
}

// Delete removes a service from the database
func (s *Service) Delete(ctx context.Context, service *models.Service) error {
	_, err := s.DB.GetDB(ctx).ModelContext(ctx, service).WherePK().Delete()
	if err != nil {
		return err
	}

	return s.Bus.Publish(ctx, &models.ServiceDeletedEvent{
		ServiceID: service.ServiceID,
	})
}

// FindStreamPermissionsForService finds every stream the service has permissions for
func (s *Service) FindStreamPermissionsForService(ctx context.Context, serviceID uuid.UUID) []*models.PermissionsServicesStreams {
	var perms []*models.PermissionsServicesStreams
	err := s.DB.GetDB(ctx).ModelContext(ctx, &perms).
		Column(
			"permissions_services_streams.*",
			"Stream",
			"Stream.Project.project_id",
			"Stream.Project.name",
			"Stream.Project.Organization.organization_id",
			"Stream.Project.Organization.name",
		).
		Where("service_id = ?", serviceID).
		Order("stream__project__organization.name", "stream__project.name", "stream.name").
		Limit(200).
		Select()
	if err != nil {
		panic(err)
	}
	return perms
}
