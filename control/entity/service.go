package entity

import (
	"context"
	"regexp"
	"time"

	uuid "github.com/satori/go.uuid"
	"gopkg.in/go-playground/validator.v9"

	"gitlab.com/beneath-hq/beneath/hub"
)

// Service represents external service keys, models, and, in the future, charts, all of which need OrganizationIDs for billing
type Service struct {
	ServiceID   uuid.UUID `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	Name        string    `sql:",notnull",validate:"required,gte=1,lte=40"` // not unique because of (organization_id, service_id) index
	Description string    `validate:"omitempty,lte=255"`
	SourceURL   string    `validate:"omitempty,url,lte=255"`
	ProjectID   uuid.UUID `sql:"on_delete:restrict,notnull,type:uuid"`
	Project     *Project
	ReadQuota   *int64    // NOTE: when updating value, clear secret cache
	WriteQuota  *int64    // NOTE: when updating value, clear secret cache
	ScanQuota   *int64    // NOTE: when updating value, clear secret cache
	CreatedOn   time.Time `sql:",notnull,default:now()"`
	UpdatedOn   time.Time `sql:",notnull,default:now()"`
	Secrets     []*ServiceSecret
}

var (
	// used for validation
	serviceNameRegex *regexp.Regexp
)

func init() {
	// configure validation
	serviceNameRegex = regexp.MustCompile("^[_a-z][_a-z0-9]*$")
	GetValidator().RegisterStructValidation(serviceValidation, Service{})
}

// custom service validation
func serviceValidation(sl validator.StructLevel) {
	s := sl.Current().Interface().(Service)

	if !serviceNameRegex.MatchString(s.Name) {
		sl.ReportError(s.Name, "Name", "", "alphanumericorunderscore", "")
	}
}

// FindService returns the matching service or nil
func FindService(ctx context.Context, serviceID uuid.UUID) *Service {
	service := &Service{
		ServiceID: serviceID,
	}
	err := hub.DB.ModelContext(ctx, service).
		Column(
			"service.*",
			"Project",
			"Project.Organization",
		).
		WherePK().
		Select()
	if !AssertFoundOne(err) {
		return nil
	}

	return service
}

// FindServiceByOrganizationProjectAndName returns the matching service or nil
func FindServiceByOrganizationProjectAndName(ctx context.Context, organizationName, projectName, serviceName string) *Service {
	service := &Service{}
	err := hub.DB.ModelContext(ctx, service).
		Column(
			"service.*",
			"Project",
			"Project.Organization",
		).
		Where("lower(project__organization.name) = lower(?)", organizationName).
		Where("lower(project.name) = lower(?)", projectName).
		Where("lower(service.name) = lower(?)", serviceName).
		Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return service
}

// Stage creates or updates the service
func (s *Service) Stage(ctx context.Context, description *string, sourceURL *string, readQuota *int64, writeQuota *int64, scanQuota *int64) error {
	// determine whether to insert or update
	update := (s.ServiceID != uuid.Nil)

	// tracks whether a save is necessary
	save := !update

	if description != nil {
		if s.Description != *description {
			s.Description = *description
			save = true
		}
	}

	if sourceURL != nil {
		if s.SourceURL != *sourceURL {
			s.SourceURL = *sourceURL
			save = true
		}
	}

	if readQuota != nil {
		if *readQuota == 0 {
			s.ReadQuota = nil
		} else {
			s.ReadQuota = readQuota
		}
		save = true
	}

	if writeQuota != nil {
		if *writeQuota == 0 {
			s.WriteQuota = nil
		} else {
			s.WriteQuota = writeQuota
		}
		save = true
	}

	if scanQuota != nil {
		if *scanQuota == 0 {
			s.ScanQuota = nil
		} else {
			s.ScanQuota = scanQuota
		}
		save = true
	}

	// quit if no changes
	if !save {
		return nil
	}

	// validate
	err := GetValidator().Struct(s)
	if err != nil {
		return err
	}

	// populate s.Project and s.Project.Organization if not set (i.e. due to inserting new stream)
	if s.Project == nil {
		s.Project = FindProject(ctx, s.ProjectID)
	}

	if update {
		s.UpdatedOn = time.Now()
		_, err := hub.DB.ModelContext(ctx, s).WherePK().Update()
		if err != nil {
			return err
		}

		getSecretCache().ClearForService(ctx, s.ServiceID)
	} else {
		_, err := hub.DB.ModelContext(ctx, s).Insert()
		if err != nil {
			return err
		}
	}

	return nil
}

// Delete removes a service from the database
func (s *Service) Delete(ctx context.Context) error {
	_, err := hub.DB.ModelContext(ctx, s).WherePK().Delete()
	return err
}
