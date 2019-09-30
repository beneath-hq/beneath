package entity

import (
	"context"
	"time"

	"github.com/beneath-core/beneath-go/core/log"
	"github.com/beneath-core/beneath-go/db"
	uuid "github.com/satori/go.uuid"
)

// Service represents external service keys, models, and, in the future, charts, all of which need OrganizationIDs for billing
type Service struct {
	ServiceID      uuid.UUID `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	Name           string
	Kind           ServiceKind `sql:",notnull"`
	OrganizationID uuid.UUID   `sql:"on_delete:cascade,notnull,type:uuid"`
	Organization   *Organization
	ReadQuota      int64
	WriteQuota     int64
	CreatedOn      time.Time `sql:",notnull,default:now()"`
	UpdatedOn      time.Time `sql:",notnull,default:now()"`
	Secrets        []*Secret
}

// ServiceKind represents a external, model, etc. services
type ServiceKind string

const (
	// ServiceKindExternal is a service key that people can put on external facing products
	ServiceKindExternal ServiceKind = "external"

	// ServiceKindModel is a service key for models
	ServiceKindModel ServiceKind = "model"
)

// FindService returns the matching service or nil
func FindService(ctx context.Context, serviceID uuid.UUID) *Service {
	service := &Service{
		ServiceID: serviceID,
	}
	err := db.DB.ModelContext(ctx, service).WherePK().Select()
	if !AssertFoundOne(err) {
		return nil
	}

	return service
}

// CreateService consolidates and returns the service matching the args
func CreateService(ctx context.Context, name string, kind ServiceKind, organizationID uuid.UUID, readBytesQuota int, writeBytesQuota int) (*Service, error) {
	s := &Service{}

	// set service fields
	s.Name = name
	s.Kind = kind
	s.OrganizationID = organizationID
	s.ReadQuota = int64(readBytesQuota)
	s.WriteQuota = int64(writeBytesQuota)

	// validate
	err := GetValidator().Struct(s)
	if err != nil {
		return nil, err
	}

	// insert
	err = db.DB.Insert(s)
	if err != nil {
		return nil, err
	}

	log.S.Infow(
		"control created service",
		"service_id", s.ServiceID,
	)

	return s, nil
}

// UpdateDetails consolidates and returns the service matching the args
func (s *Service) UpdateDetails(ctx context.Context, name *string, organizationID *uuid.UUID, readBytesQuota *int, writeBytesQuota *int) (*Service, error) {
	// set fields
	if name != nil {
		s.Name = *name
	}
	if organizationID != nil {
		s.OrganizationID = *organizationID
	}
	if readBytesQuota != nil {
		s.ReadQuota = int64(*readBytesQuota)
	}
	if writeBytesQuota != nil {
		s.WriteQuota = int64(*writeBytesQuota)
	}

	// validate
	err := GetValidator().Struct(s)
	if err != nil {
		return nil, err
	}

	// update
	s.UpdatedOn = time.Now()
	_, err = db.DB.ModelContext(ctx, s).Column("name", "organization_id", "read_quota", "write_quota", "updated_on").WherePK().Update()
	return s, err
}

// Delete removes a service from the database
func (s *Service) Delete(ctx context.Context) error {
	_, err := db.DB.ModelContext(ctx, s).WherePK().Delete()
	return err
}

// UpdatePermissions updates a service's permissions for a given stream
// UpdatePermissions sets permissions if they do not exist yet
func (s *Service) UpdatePermissions(ctx context.Context, streamID uuid.UUID, read bool, write bool) (*PermissionsServicesStreams, error) {
	// change data
	pss := &PermissionsServicesStreams{
		ServiceID: s.ServiceID,
		StreamID:  streamID,
		Read:      read,
		Write:     write,
	}

	// validate
	err := GetValidator().Struct(pss)
	if err != nil {
		return nil, err
	}

	// update
	_, err = db.DB.ModelContext(ctx, pss).
		OnConflict("(service_id, stream_id) DO UPDATE").
		Set("read = EXCLUDED.read").
		Set("write = EXCLUDED.write").
		Insert()
	return pss, err
}
