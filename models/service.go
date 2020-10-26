package models

import (
	"regexp"
	"time"

	uuid "github.com/satori/go.uuid"
	"gopkg.in/go-playground/validator.v9"
)

// Service represents non-user accounts, which have distinct access permissions, API secrets, monitoring and quotas.
// They're used eg. when deploying to production.
type Service struct {
	_msgpack    struct{}         `msgpack:",omitempty"`
	ServiceID   uuid.UUID        `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	Name        string           `sql:",notnull",validate:"required,gte=1,lte=40"` // not unique because of (organization_id, service_id) index
	Description string           `validate:"omitempty,lte=255"`
	SourceURL   string           `validate:"omitempty,url,lte=255"`
	ProjectID   uuid.UUID        `sql:"on_delete:restrict,notnull,type:uuid"`
	Project     *Project         `msgpack:"-"`
	ReadQuota   *int64           // NOTE: when updating value, clear secret cache
	WriteQuota  *int64           // NOTE: when updating value, clear secret cache
	ScanQuota   *int64           // NOTE: when updating value, clear secret cache
	CreatedOn   time.Time        `sql:",notnull,default:now()"`
	UpdatedOn   time.Time        `sql:",notnull,default:now()"`
	Secrets     []*ServiceSecret `msgpack:"-"`
}

// Validate runs validation on the service
func (s *Service) Validate() error {
	return Validator.Struct(s)
}

// ---------------
// Events

// ServiceUpdatedEvent is sent when a service is updated
type ServiceUpdatedEvent struct {
	Service *Service
}

// ServiceDeletedEvent is sent when a service is deleted
type ServiceDeletedEvent struct {
	ServiceID uuid.UUID
}

// ---------------
// Validation

var serviceNameRegex = regexp.MustCompile("^[_a-z][_a-z0-9]*$")

func init() {
	Validator.RegisterStructValidation(func(sl validator.StructLevel) {
		s := sl.Current().Interface().(Service)
		if !serviceNameRegex.MatchString(s.Name) {
			sl.ReportError(s.Name, "Name", "", "alphanumericorunderscore", "")
		}
	}, Service{})
}
