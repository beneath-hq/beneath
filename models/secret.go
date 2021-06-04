package models

import (
	"fmt"
	"time"

	uuid "github.com/satori/go.uuid"

	"github.com/beneath-hq/beneath/pkg/secrettoken"
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

	// GetBillingQuotaEpoch returns the starting time for the current billing quotas
	GetBillingQuotaEpoch() time.Time

	// GetBillingReadQuota returns the billing organization's read quota (or nil if unlimited)
	GetBillingReadQuota() *int64

	// GetBillingWriteQuota returns the billing organization's write quota (or nil if unlimited)
	GetBillingWriteQuota() *int64

	// GetBillingScanQuota returns the billing organization's scan quota (or nil if unlimited)
	GetBillingScanQuota() *int64

	// GetOwnerQuotaEpoch returns the starting time for the current owner quotas
	GetOwnerQuotaEpoch() time.Time

	// GetOwnerReadQuota returns the owner's read quota (or nil if unlimited)
	GetOwnerReadQuota() *int64

	// GetOwnerWriteQuota returns the owner's write quota (or nil if unlimited)
	GetOwnerWriteQuota() *int64

	// GetOwnerScanQuota returns the owner's scan quota (or nil if unlimited)
	GetOwnerScanQuota() *int64
}

// BaseSecret is the "abstract" base of structs that implement the Secret interface
type BaseSecret struct {
	_msgpack    struct{}  `msgpack:",omitempty"`
	Prefix      string    `sql:",notnull",validate:"required,len=4"`
	HashedToken []byte    `sql:",unique,notnull",validate:"required,lte=64"`
	Description string    `validate:"omitempty,lte=40"`
	CreatedOn   time.Time `sql:",notnull,default:now()"`
	UpdatedOn   time.Time `sql:",notnull,default:now()"`

	Token                 secrettoken.Token `sql:"-"`
	Master                bool              `sql:"-"`
	BillingOrganizationID uuid.UUID         `sql:"-"`
	BillingQuotaEpoch     time.Time         `sql:"-"`
	BillingReadQuota      *int64            `sql:"-"`
	BillingWriteQuota     *int64            `sql:"-"`
	BillingScanQuota      *int64            `sql:"-"`
	OwnerQuotaEpoch       time.Time         `sql:"-"`
	OwnerReadQuota        *int64            `sql:"-"`
	OwnerWriteQuota       *int64            `sql:"-"`
	OwnerScanQuota        *int64            `sql:"-"`
}

// GetBillingOrganizationID implements Secret
func (s *BaseSecret) GetBillingOrganizationID() uuid.UUID {
	return s.BillingOrganizationID
}

// IsMaster implements Secret
func (s *BaseSecret) IsMaster() bool {
	return s.Master
}

// GetBillingQuotaEpoch implements Secret
func (s *BaseSecret) GetBillingQuotaEpoch() time.Time {
	return s.BillingQuotaEpoch
}

// GetBillingReadQuota implements Secret
func (s *BaseSecret) GetBillingReadQuota() *int64 {
	return s.BillingReadQuota
}

// GetBillingWriteQuota implements Secret
func (s *BaseSecret) GetBillingWriteQuota() *int64 {
	return s.BillingWriteQuota
}

// GetBillingScanQuota implements Secret
func (s *BaseSecret) GetBillingScanQuota() *int64 {
	return s.BillingScanQuota
}

// GetOwnerQuotaEpoch implements Secret
func (s *BaseSecret) GetOwnerQuotaEpoch() time.Time {
	return s.OwnerQuotaEpoch
}

// GetOwnerReadQuota implements Secret
func (s *BaseSecret) GetOwnerReadQuota() *int64 {
	return s.OwnerReadQuota
}

// GetOwnerWriteQuota implements Secret
func (s *BaseSecret) GetOwnerWriteQuota() *int64 {
	return s.OwnerWriteQuota
}

// GetOwnerScanQuota implements Secret
func (s *BaseSecret) GetOwnerScanQuota() *int64 {
	return s.OwnerScanQuota
}

// UserSecret implements Secret for User entities
type UserSecret struct {
	_msgpack     struct{}  `msgpack:",omitempty"`
	UserSecretID uuid.UUID `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	BaseSecret
	UserID     uuid.UUID `sql:"on_delete:CASCADE,notnull,type:uuid"`
	User       *User     `msgpack:"-"`
	ReadOnly   bool      `sql:",notnull"`
	PublicOnly bool      `sql:",notnull"`
}

// Validate runs validation on the secret
func (s *UserSecret) Validate() error {
	return Validator.Struct(s)
}

// GetSecretID implements the Secret interface
func (s *UserSecret) GetSecretID() uuid.UUID {
	return s.UserSecretID
}

// GetOwnerID implements the Secret interface
func (s *UserSecret) GetOwnerID() uuid.UUID {
	return s.UserID
}

// IsAnonymous implements the Secret interface
func (s *UserSecret) IsAnonymous() bool {
	return false
}

// IsUser implements the Secret interface
func (s *UserSecret) IsUser() bool {
	return true
}

// IsService implements the Secret interface
func (s *UserSecret) IsService() bool {
	return false
}

// ServiceSecret implements Secret for Token entities
type ServiceSecret struct {
	_msgpack        struct{}  `msgpack:",omitempty"`
	ServiceSecretID uuid.UUID `sql:",pk,type:uuid,default:uuid_generate_v4()"`
	BaseSecret
	ServiceID uuid.UUID `sql:"on_delete:CASCADE,type:uuid"`
	Service   *Service  `msgpack:"-"`
}

// Validate runs validation on the secret
func (s *ServiceSecret) Validate() error {
	return Validator.Struct(s)
}

// GetSecretID implements the Secret interface
func (s *ServiceSecret) GetSecretID() uuid.UUID {
	return s.ServiceSecretID
}

// GetOwnerID implements the Secret interface
func (s *ServiceSecret) GetOwnerID() uuid.UUID {
	return s.ServiceID
}

// IsAnonymous implements the Secret interface
func (s *ServiceSecret) IsAnonymous() bool {
	return false
}

// IsUser implements the Secret interface
func (s *ServiceSecret) IsUser() bool {
	return false
}

// IsService implements the Secret interface
func (s *ServiceSecret) IsService() bool {
	return true
}

// IsMaster implements the Secret interface
func (s *ServiceSecret) IsMaster() bool {
	return false
}

// AnonymousSecret implements Secret for anonymous requests
type AnonymousSecret struct{}

// GetSecretID implements the Secret interface
func (s *AnonymousSecret) GetSecretID() uuid.UUID {
	return uuid.Nil
}

// GetOwnerID implements the Secret interface
func (s *AnonymousSecret) GetOwnerID() uuid.UUID {
	return uuid.Nil
}

// GetBillingOrganizationID implements the Secret interface
func (s *AnonymousSecret) GetBillingOrganizationID() uuid.UUID {
	return uuid.Nil
}

// IsAnonymous implements the Secret interface
func (s *AnonymousSecret) IsAnonymous() bool {
	return true
}

// IsUser implements the Secret interface
func (s *AnonymousSecret) IsUser() bool {
	return false
}

// IsService implements the Secret interface
func (s *AnonymousSecret) IsService() bool {
	return false
}

// IsMaster implements the Secret interface
func (s *AnonymousSecret) IsMaster() bool {
	return false
}

// GetBillingQuotaEpoch implements Secret
func (s *AnonymousSecret) GetBillingQuotaEpoch() time.Time {
	panic(fmt.Errorf("Called GetBillingQuotaEpoch on an anonymous secret"))
}

// GetBillingReadQuota implements Secret
func (s *AnonymousSecret) GetBillingReadQuota() *int64 {
	panic(fmt.Errorf("Called GetBillingReadQuota on an anonymous secret"))
}

// GetBillingWriteQuota implements Secret
func (s *AnonymousSecret) GetBillingWriteQuota() *int64 {
	panic(fmt.Errorf("Called GetBillingWriteQuota on an anonymous secret"))
}

// GetBillingScanQuota implements Secret
func (s *AnonymousSecret) GetBillingScanQuota() *int64 {
	panic(fmt.Errorf("Called GetBillingScanQuota on an anonymous secret"))
}

// GetOwnerQuotaEpoch implements Secret
func (s *AnonymousSecret) GetOwnerQuotaEpoch() time.Time {
	panic(fmt.Errorf("Called GetOwnerQuotaEpoch on an anonymous secret"))
}

// GetOwnerReadQuota implements Secret
func (s *AnonymousSecret) GetOwnerReadQuota() *int64 {
	panic(fmt.Errorf("Called GetOwnerReadQuota on an anonymous secret"))
}

// GetOwnerWriteQuota implements Secret
func (s *AnonymousSecret) GetOwnerWriteQuota() *int64 {
	panic(fmt.Errorf("Called GetOwnerWriteQuota on an anonymous secret"))
}

// GetOwnerScanQuota implements Secret
func (s *AnonymousSecret) GetOwnerScanQuota() *int64 {
	panic(fmt.Errorf("Called GetOwnerScanQuota on an anonymous secret"))
}
