package entity

import (
	"context"
	"regexp"
	"time"

	uuid "github.com/satori/go.uuid"
	"gitlab.com/beneath-hq/beneath/internal/hub"
	"gopkg.in/go-playground/validator.v9"
)

// Service represents external service keys, models, and, in the future, charts, all of which need OrganizationIDs for billing
type Service struct {
	ServiceID      uuid.UUID   `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	Name           string      `sql:",notnull",validate:"required,gte=1,lte=40"` // not unique because of (organization_id, service_id) index
	Kind           ServiceKind `sql:",notnull"`
	OrganizationID uuid.UUID   `sql:"on_delete:restrict,notnull,type:uuid"`
	Organization   *Organization
	ReadQuota      *int64
	WriteQuota     *int64
	CreatedOn      time.Time `sql:",notnull,default:now()"`
	UpdatedOn      time.Time `sql:",notnull,default:now()"`
	Secrets        []*ServiceSecret
}

// ServiceKind represents a external, model, etc. services
type ServiceKind string

const (
	// ServiceKindExternal is a service key that people can put on external facing products
	ServiceKindExternal ServiceKind = "external"

	// ServiceKindModel is a service key for models
	ServiceKindModel ServiceKind = "model"
)

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
	err := hub.DB.ModelContext(ctx, service).WherePK().Select()
	if !AssertFoundOne(err) {
		return nil
	}

	return service
}

// FindServiceByNameAndOrganization returns the matching service or nil
func FindServiceByNameAndOrganization(ctx context.Context, name, organizationName string) *Service {
	service := &Service{}
	err := hub.DB.ModelContext(ctx, service).
		Column("service.*", "Organization").
		Where("lower(service.name) = lower(?)", name).
		Where("lower(organization.name) = lower(?)", organizationName).
		Select()
	if !AssertFoundOne(err) {
		return nil
	}
	return service
}

// CreateService consolidates and returns the service matching the args
func CreateService(ctx context.Context, name string, kind ServiceKind, organizationID uuid.UUID, readQuota *int64, writeQuota *int64) (*Service, error) {
	s := &Service{}

	// set service fields
	s.Name = name
	s.Kind = kind
	s.OrganizationID = organizationID
	s.ReadQuota = readQuota
	s.WriteQuota = writeQuota

	// validate
	err := GetValidator().Struct(s)
	if err != nil {
		return nil, err
	}

	// insert
	err = hub.DB.Insert(s)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// UpdateDetails consolidates and returns the service matching the args
func (s *Service) UpdateDetails(ctx context.Context, name *string) (*Service, error) {
	// set fields
	if name != nil {
		s.Name = *name
	}

	// validate
	err := GetValidator().Struct(s)
	if err != nil {
		return nil, err
	}

	// update
	s.UpdatedOn = time.Now()
	_, err = hub.DB.ModelContext(ctx, s).Column("name", "updated_on").WherePK().Update()
	return s, err
}

// UpdateQuotas updates the quotas enforced upon the service
func (s *Service) UpdateQuotas(ctx context.Context, readQuota *int64, writeQuota *int64) error {
	// set fields
	s.ReadQuota = readQuota
	s.WriteQuota = writeQuota

	// validate
	err := GetValidator().Struct(s)
	if err != nil {
		return err
	}

	// update
	s.UpdatedOn = time.Now()
	_, err = hub.DB.ModelContext(ctx, s).Column("read_quota", "write_quota", "updated_on").WherePK().Update()
	return err
}

// Delete removes a service from the database
func (s *Service) Delete(ctx context.Context) error {
	_, err := hub.DB.ModelContext(ctx, s).WherePK().Delete()
	return err
}
