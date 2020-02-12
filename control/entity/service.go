package entity

import (
	"context"
	"regexp"
	"time"

	"github.com/beneath-core/internal/hub"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/go-playground/validator.v9"
)

// Service represents external service keys, models, and, in the future, charts, all of which need OrganizationIDs for billing
type Service struct {
	ServiceID      uuid.UUID   `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	Name           string      `sql:",notnull",validate:"required,gte=1,lte=40"` // not unique because of (organization_id, service_id) index
	Kind           ServiceKind `sql:",notnull"`
	OrganizationID uuid.UUID   `sql:"on_delete:restrict,notnull,type:uuid"`
	Organization   *Organization
	ReadQuota      int64
	WriteQuota     int64
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
func CreateService(ctx context.Context, name string, kind ServiceKind, organizationID uuid.UUID, readQuota int, writeQuota int) (*Service, error) {
	s := &Service{}

	// set service fields
	s.Name = name
	s.Kind = kind
	s.OrganizationID = organizationID
	s.ReadQuota = int64(readQuota)
	s.WriteQuota = int64(writeQuota)

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
func (s *Service) UpdateDetails(ctx context.Context, name *string, readQuota *int, writeQuota *int) (*Service, error) {
	// set fields
	if name != nil {
		s.Name = *name
	}
	if readQuota != nil {
		s.ReadQuota = int64(*readQuota)
	}
	if writeQuota != nil {
		s.WriteQuota = int64(*writeQuota)
	}

	// validate
	err := GetValidator().Struct(s)
	if err != nil {
		return nil, err
	}

	// update
	s.UpdatedOn = time.Now()
	_, err = hub.DB.ModelContext(ctx, s).Column("name", "read_quota", "write_quota", "updated_on").WherePK().Update()
	return s, err
}

// UpdateOrganization changes a service's organization
func (s *Service) UpdateOrganization(ctx context.Context, organizationID uuid.UUID) (*Service, error) {
	s.OrganizationID = organizationID
	s.UpdatedOn = time.Now()

	// commit current usage to the old organization's bill
	err := commitCurrentUsageToNextBill(ctx, s.OrganizationID, ServiceEntityKind, s.ServiceID, s.Name, false)
	if err != nil {
		return nil, err
	}

	_, err = hub.DB.ModelContext(ctx, s).Column("organization_id", "updated_on").WherePK().Update()

	// commit usage credit to the new organization's bill for the service's current month's usage
	err = commitCurrentUsageToNextBill(ctx, organizationID, ServiceEntityKind, s.ServiceID, s.Name, true)
	if err != nil {
		panic("unable to commit usage credit to bill")
	}

	// re-find service so we can return the name of the new organization
	err = hub.DB.ModelContext(ctx, s).
		WherePK().
		Column("service.*", "Organization").
		Select()
	if !AssertFoundOne(err) {
		return nil, err
	}

	return s, err
}

// Delete removes a service from the database
func (s *Service) Delete(ctx context.Context) error {
	err := commitCurrentUsageToNextBill(ctx, s.OrganizationID, ServiceEntityKind, s.ServiceID, s.Name, false)
	if err != nil {
		return err
	}
	_, err = hub.DB.ModelContext(ctx, s).WherePK().Delete()
	return err
}

// UpdatePermissions updates a service's permissions for a given stream
// UpdatePermissions sets permissions if they do not exist yet
func (s *Service) UpdatePermissions(ctx context.Context, streamID uuid.UUID, read *bool, write *bool) (*PermissionsServicesStreams, error) {
	// create perm
	pss := &PermissionsServicesStreams{
		ServiceID: s.ServiceID,
		StreamID:  streamID,
	}
	if read != nil {
		pss.Read = *read
	}
	if write != nil {
		pss.Write = *write
	}

	// if neither read nor write, delete permission (if exists) -- else update
	if !pss.Read && !pss.Write {
		_, err := hub.DB.ModelContext(ctx, pss).WherePK().Delete()
		if err != nil {
			return nil, err
		}
	} else {
		// build upsert
		q := hub.DB.ModelContext(ctx, pss).OnConflict("(service_id, stream_id) DO UPDATE")
		if read != nil {
			q = q.Set("read = EXCLUDED.read")
		}
		if write != nil {
			q = q.Set("write = EXCLUDED.write")
		}

		// run upsert
		_, err := q.Insert()
		if err != nil {
			return nil, err
		}
	}

	return pss, nil
}
