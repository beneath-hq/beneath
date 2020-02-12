package entity

import (
	"context"

	"github.com/beneath-core/pkg/log"

	pb "github.com/beneath-core/engine/proto"
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

// CheckReadQuota implements Secret
func (s *AnonymousSecret) CheckReadQuota(u pb.QuotaUsage) bool {
	log.S.Warnf("called CheckReadQuota on AnonymousSecret")
	return true
}

// CheckWriteQuota implements Secret
func (s *AnonymousSecret) CheckWriteQuota(u pb.QuotaUsage) bool {
	log.S.Warnf("called CheckWriteQuota on AnonymousSecret")
	return true
}

// StreamPermissions implements the Secret interface
func (s *AnonymousSecret) StreamPermissions(ctx context.Context, streamID uuid.UUID, projectID uuid.UUID, public bool, external bool) StreamPermissions {
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

// ManagesModelBatches implements the Secret interface
func (s *AnonymousSecret) ManagesModelBatches(model *Model) bool {
	return false
}

// Revoke implements the Secret interface
func (s *AnonymousSecret) Revoke(ctx context.Context) {
	log.S.Warnf("called Revoke on AnonymousSecret")
}
