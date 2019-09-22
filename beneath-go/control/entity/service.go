package entity

import (
	"context"

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
	Secrets        []*Secret
}

// ServiceKind represents a external, model, etc. services
type ServiceKind string

const (
	// ServiceKindExternal is a service key that people can put on external facing products
	ServiceKindExternal ServiceKind = "external"

	// ServiceKindModel is a service key for models
	ServiceKindModel ServiceKind = "model"

	// DefaultServiceReadQuota is the default read quota for service keys
	DefaultServiceReadQuota = 100000000

	// DefaultServiceWriteQuota is the default write quota for service keys
	DefaultServiceWriteQuota = 100000000
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
