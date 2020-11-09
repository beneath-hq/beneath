package service

import (
	"context"
	"time"

	uuid "github.com/satori/go.uuid"
	"gitlab.com/beneath-hq/beneath/bus"
	"gitlab.com/beneath-hq/beneath/infra/db"
	"gitlab.com/beneath-hq/beneath/models"
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

// Stage creates or updates the service
func (s *Service) Stage(ctx context.Context, service *models.Service, description *string, sourceURL *string, readQuota *int64, writeQuota *int64, scanQuota *int64) error {
	// determine whether to insert or update
	update := (service.ServiceID != uuid.Nil)

	// tracks whether a save is necessary
	save := !update

	if description != nil {
		if service.Description != *description {
			service.Description = *description
			save = true
		}
	}

	if sourceURL != nil {
		if service.SourceURL != *sourceURL {
			service.SourceURL = *sourceURL
			save = true
		}
	}

	if readQuota != nil {
		if *readQuota == 0 {
			service.ReadQuota = nil
		} else {
			service.ReadQuota = readQuota
		}
		save = true
	}

	if writeQuota != nil {
		if *writeQuota == 0 {
			service.WriteQuota = nil
		} else {
			service.WriteQuota = writeQuota
		}
		save = true
	}

	if scanQuota != nil {
		if *scanQuota == 0 {
			service.ScanQuota = nil
		} else {
			service.ScanQuota = scanQuota
		}
		save = true
	}

	// quit if no changes
	if !save {
		return nil
	}

	// validate
	err := service.Validate()
	if err != nil {
		return err
	}

	if update {
		service.UpdatedOn = time.Now()
		_, err := s.DB.GetDB(ctx).ModelContext(ctx, service).WherePK().Update()
		if err != nil {
			return err
		}

		err = s.Bus.Publish(ctx, &models.ServiceUpdatedEvent{
			Service: service,
		})
		if err != nil {
			return err
		}
	} else {
		_, err := s.DB.GetDB(ctx).ModelContext(ctx, service).Insert()
		if err != nil {
			return err
		}
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
