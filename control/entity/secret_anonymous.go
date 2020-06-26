package entity

import (
	"context"
	"fmt"

	"gitlab.com/beneath-hq/beneath/pkg/log"

	uuid "github.com/satori/go.uuid"
)

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

// GetBillingReadQuota implements Secret
func (s *AnonymousSecret) GetBillingReadQuota() *int64 {
	panic(fmt.Errorf("Called GetBillingReadQuota on an anonymous secret"))
}

// GetBillingWriteQuota implements Secret
func (s *AnonymousSecret) GetBillingWriteQuota() *int64 {
	panic(fmt.Errorf("Called GetBillingWriteQuota on an anonymous secret"))
}

// GetOwnerReadQuota implements Secret
func (s *AnonymousSecret) GetOwnerReadQuota() *int64 {
	panic(fmt.Errorf("Called GetOwnerReadQuota on an anonymous secret"))
}

// GetOwnerWriteQuota implements Secret
func (s *AnonymousSecret) GetOwnerWriteQuota() *int64 {
	panic(fmt.Errorf("Called GetOwnerWriteQuota on an anonymous secret"))
}

// StreamPermissions implements the Secret interface
func (s *AnonymousSecret) StreamPermissions(ctx context.Context, streamID uuid.UUID, projectID uuid.UUID, public bool) StreamPermissions {
	return StreamPermissions{
		Read: public,
	}
}

// ProjectPermissions implements the Secret interface
func (s *AnonymousSecret) ProjectPermissions(ctx context.Context, projectID uuid.UUID, public bool) ProjectPermissions {
	return ProjectPermissions{
		View: public,
	}
}

// OrganizationPermissions implements the Secret interface
func (s *AnonymousSecret) OrganizationPermissions(ctx context.Context, organizationID uuid.UUID) OrganizationPermissions {
	return OrganizationPermissions{}
}

// Revoke implements the Secret interface
func (s *AnonymousSecret) Revoke(ctx context.Context) {
	log.S.Warnf("called Revoke on AnonymousSecret")
}
