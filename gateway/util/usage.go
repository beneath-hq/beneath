package util

import (
	"context"
	"fmt"

	uuid "github.com/satori/go.uuid"

	"gitlab.com/beneath-hq/beneath/control/entity"
	"gitlab.com/beneath-hq/beneath/gateway"
)

const (
	minimumBytesBilled = 1024
)

// TrackRead is a helper to track read usage for a secret from an instance
func TrackRead(ctx context.Context, secret entity.Secret, streamID uuid.UUID, instanceID uuid.UUID, nrecords int64, nbytes int64) {
	gateway.Metrics.TrackRead(streamID, nrecords, nbytes)
	gateway.Metrics.TrackRead(instanceID, nrecords, nbytes)
	if !secret.IsAnonymous() {
		if nbytes < minimumBytesBilled {
			nbytes = minimumBytesBilled
		}
		gateway.Metrics.TrackRead(secret.GetOwnerID(), nrecords, nbytes)
		gateway.Metrics.TrackRead(secret.GetBillingOrganizationID(), nrecords, nbytes)
	}
}

// TrackWrite is a helper to track write usage for a secret to an instance
func TrackWrite(ctx context.Context, secret entity.Secret, streamID uuid.UUID, instanceID uuid.UUID, nrecords int64, nbytes int64) {
	gateway.Metrics.TrackWrite(streamID, nrecords, nbytes)
	gateway.Metrics.TrackWrite(instanceID, nrecords, nbytes)
	if !secret.IsAnonymous() {
		if nbytes < minimumBytesBilled {
			nbytes = minimumBytesBilled
		}
		gateway.Metrics.TrackWrite(secret.GetOwnerID(), nrecords, nbytes)
		gateway.Metrics.TrackWrite(secret.GetBillingOrganizationID(), nrecords, nbytes)
	}
}

// CheckReadQuota checks that the secret is within its quotas to trigger a read
func CheckReadQuota(ctx context.Context, secret entity.Secret) error {
	if secret.IsAnonymous() {
		return nil
	}

	brq := secret.GetBillingReadQuota()
	if brq != nil {
		usage := gateway.Metrics.GetCurrentUsage(ctx, secret.GetBillingOrganizationID())
		if usage.ReadBytes >= *brq {
			return fmt.Errorf("your organization has exhausted its monthly read quota")
		}
	}

	orq := secret.GetOwnerReadQuota()
	if orq != nil {
		usage := gateway.Metrics.GetCurrentUsage(ctx, secret.GetOwnerID())
		if usage.ReadBytes >= *orq {
			return fmt.Errorf("you have exhausted your monthly read quota")
		}
	}

	return nil
}

// CheckWriteQuota checks that the secret is within its quotas to trigger a write
func CheckWriteQuota(ctx context.Context, secret entity.Secret) error {
	if secret.IsAnonymous() {
		return nil
	}

	bwq := secret.GetBillingWriteQuota()
	if bwq != nil {
		usage := gateway.Metrics.GetCurrentUsage(ctx, secret.GetBillingOrganizationID())
		if usage.WriteBytes >= *bwq {
			return fmt.Errorf("your organization has exhausted its monthly write quota")
		}
	}

	owq := secret.GetOwnerWriteQuota()
	if owq != nil {
		usage := gateway.Metrics.GetCurrentUsage(ctx, secret.GetOwnerID())
		if usage.WriteBytes >= *owq {
			return fmt.Errorf("you have exhausted your monthly write quota")
		}
	}

	return nil
}
