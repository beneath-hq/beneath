package entity

import (
	"context"
	"time"

	"github.com/beneath-core/beneath-go/core/secrettoken"

	pb "github.com/beneath-core/beneath-go/proto"
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

	// IsAnonymous is true iff the secret is anonymous
	IsAnonymous() bool

	// IsService is true iff the secret is a service
	IsService() bool

	// IsUser is true iff the secret is a user
	IsUser() bool

	// Checks if the secret owner is within its read quota
	CheckReadQuota(u pb.QuotaUsage) bool

	// Checks if the secret owner is within its write quota
	CheckWriteQuota(u pb.QuotaUsage) bool

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
	HashedToken string    `sql:",unique,notnull",validate:"required,lte=64"`
	Description string    `validate:"omitempty,lte=32"`
	CreatedOn   time.Time `sql:",notnull,default:now()"`
	UpdatedOn   time.Time `sql:",notnull,default:now()"`

	Token      secrettoken.Token `sql:"-"`
	ReadQuota  int64             `sql:"-"`
	WriteQuota int64             `sql:"-"`
}

// CheckReadQuota implements Secret
func (s *BaseSecret) CheckReadQuota(u pb.QuotaUsage) bool {
	return u.ReadBytes < s.ReadQuota
}

// CheckWriteQuota implements Secret
func (s *BaseSecret) CheckWriteQuota(u pb.QuotaUsage) bool {
	return u.WriteBytes < s.WriteQuota
}
