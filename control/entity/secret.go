package entity

import (
	"context"
	"time"

	"gitlab.com/beneath-hq/beneath/pkg/secrettoken"

	uuid "github.com/satori/go.uuid"
)

const (
	// TokenFlagsService is used as flags byte for service secret tokens
	TokenFlagsService = byte(0x80)

	// TokenFlagsUser is used as flags byte for user secret tokens
	TokenFlagsUser = byte(0x81)
)

// Secret represents an access token to Beneath
type Secret interface {
	// GetSecretID returns a unique identifier of the secret
	GetSecretID() uuid.UUID

	// GetOwnerID returns the ID of the secret's owner, i.e. a user or service (or uuid.Nil for anonymous)
	GetOwnerID() uuid.UUID

	// GetBillingOrganizationID returns the ID of the organization responsible for the secret's billing
	GetBillingOrganizationID() uuid.UUID

	// IsAnonymous is true iff the secret is anonymous
	IsAnonymous() bool

	// IsUser is true iff the secret is a user
	IsUser() bool

	// IsService is true iff the secret is a service
	IsService() bool

	// IsMaster is true iff the secret is a master secret
	IsMaster() bool

	// GetBillingReadQuota returns the billing organization's read quota (or nil if unlimited)
	GetBillingReadQuota() *int64

	// GetBillingWriteQuota returns the billing organization's write quota (or nil if unlimited)
	GetBillingWriteQuota() *int64

	// GetOwnerReadQuota returns the owner's read quota (or nil if unlimited)
	GetOwnerReadQuota() *int64

	// GetOwnerWriteQuota returns the owner's write quota (or nil if unlimited)
	GetOwnerWriteQuota() *int64

	// StreamPermissions returns the secret's permissions for a given stream
	StreamPermissions(ctx context.Context, streamID uuid.UUID, projectID uuid.UUID, public bool, external bool) StreamPermissions

	// ProjectPermissions returns the secret's permissions for a given project
	ProjectPermissions(ctx context.Context, projectID uuid.UUID, public bool) ProjectPermissions

	// OrganizationPermissions returns the secret's permissions for a given organization
	OrganizationPermissions(ctx context.Context, organizationID uuid.UUID) OrganizationPermissions

	// ManagesModelBatches is a hack for until we have ModelPermissions
	ManagesModelBatches(model *Model) bool

	// Revokes the secret
	Revoke(ctx context.Context)
}

// BaseSecret is the "abstract" base of structs that implement the Secret interface
type BaseSecret struct {
	_msgpack    struct{}  `msgpack:",omitempty"`
	Prefix      string    `sql:",notnull",validate:"required,len=4"`
	HashedToken []byte    `sql:",unique,notnull",validate:"required,lte=64"`
	Description string    `validate:"omitempty,lte=32"`
	CreatedOn   time.Time `sql:",notnull,default:now()"`
	UpdatedOn   time.Time `sql:",notnull,default:now()"`

	Token                 secrettoken.Token `sql:"-"`
	Master                bool              `sql:"-"`
	BillingOrganizationID uuid.UUID         `sql:"-"`
	BillingReadQuota      *int64            `sql:"-"`
	BillingWriteQuota     *int64            `sql:"-"`
	OwnerReadQuota        *int64            `sql:"-"`
	OwnerWriteQuota       *int64            `sql:"-"`
}

// GetBillingOrganizationID implements Secret
func (s *BaseSecret) GetBillingOrganizationID() uuid.UUID {
	return s.BillingOrganizationID
}

// IsMaster implements Secret
func (s *BaseSecret) IsMaster() bool {
	return s.Master
}

// GetBillingReadQuota implements Secret
func (s *BaseSecret) GetBillingReadQuota() *int64 {
	return s.BillingReadQuota
}

// GetBillingWriteQuota implements Secret
func (s *BaseSecret) GetBillingWriteQuota() *int64 {
	return s.BillingWriteQuota
}

// GetOwnerReadQuota implements Secret
func (s *BaseSecret) GetOwnerReadQuota() *int64 {
	return s.OwnerReadQuota
}

// GetOwnerWriteQuota implements Secret
func (s *BaseSecret) GetOwnerWriteQuota() *int64 {
	return s.OwnerWriteQuota
}
