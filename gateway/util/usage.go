package util

import (
	"context"
	"fmt"

	uuid "github.com/satori/go.uuid"

	"gitlab.com/beneath-hq/beneath/control/entity"
	"gitlab.com/beneath-hq/beneath/gateway"
)

const (
	minReadBytesBilled  = 1024
	minWriteBytesBilled = 1024
	minScanBytesBilled  = 1048576
)

// TrackRead is a helper to track read usage for a secret from an instance
func TrackRead(ctx context.Context, secret entity.Secret, streamID uuid.UUID, instanceID uuid.UUID, nrecords int64, nbytes int64) {
	if streamID != uuid.Nil {
		gateway.Metrics.TrackRead(streamID, nrecords, nbytes)
	}
	if instanceID != uuid.Nil {
		gateway.Metrics.TrackRead(instanceID, nrecords, nbytes)
	}
	if !secret.IsAnonymous() {
		if nbytes < minReadBytesBilled {
			nbytes = minReadBytesBilled
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
		if nbytes < minWriteBytesBilled {
			nbytes = minWriteBytesBilled
		}
		gateway.Metrics.TrackWrite(secret.GetOwnerID(), nrecords, nbytes)
		gateway.Metrics.TrackWrite(secret.GetBillingOrganizationID(), nrecords, nbytes)
	}
}

// TrackScan is a helper to track scan usage for a secret to an instance
func TrackScan(ctx context.Context, secret entity.Secret, nbytes int64) {
	if !secret.IsAnonymous() {
		if nbytes < minScanBytesBilled {
			nbytes = minScanBytesBilled
		}
		gateway.Metrics.TrackScan(secret.GetOwnerID(), nbytes)
		gateway.Metrics.TrackScan(secret.GetBillingOrganizationID(), nbytes)
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

// CheckScanQuota checks that secret is within its quotas to trigger a warehouse query
func CheckScanQuota(ctx context.Context, secret entity.Secret, estimatedScanBytes int64) error {
	if secret.IsAnonymous() {
		return fmt.Errorf("anonymous users cannot run warehouse queries")
	}

	bsq := secret.GetBillingScanQuota()
	if bsq != nil {
		usage := gateway.Metrics.GetCurrentUsage(ctx, secret.GetBillingOrganizationID())
		if usage.ScanBytes >= *bsq {
			return fmt.Errorf("your organization has exhausted its monthly warehouse scan quota")
		} else if usage.ScanBytes+estimatedScanBytes > *bsq {
			return fmt.Errorf("your organization doesn't have a sufficient remaining scan quota to execute query (estimated at %d scanned bytes)", estimatedScanBytes)
		}
	}

	osq := secret.GetOwnerScanQuota()
	if osq != nil {
		usage := gateway.Metrics.GetCurrentUsage(ctx, secret.GetOwnerID())
		if usage.ScanBytes >= *osq {
			return fmt.Errorf("you have exhausted your monthly warehouse scan quota")
		} else if usage.ScanBytes+estimatedScanBytes > *osq {
			return fmt.Errorf("your don't have a sufficient remaining scan quota to execute query (estimated at %d scanned bytes)", estimatedScanBytes)
		}
	}

	return nil
}
