package usage

import (
	"context"
	"fmt"
	"time"

	"github.com/bluele/gcache"
	uuid "github.com/satori/go.uuid"

	"github.com/beneath-hq/beneath/infra/engine/driver"
	pb "github.com/beneath-hq/beneath/infra/engine/proto"
	"github.com/beneath-hq/beneath/models"
	"github.com/beneath-hq/beneath/pkg/bytesutil"
)

// GetCurrentQuotaUsage returns the owner's quota usage in the quota period that currently applies to it.
// If the background writer is running, results will include buffered usage, and will also be cached.
// If the background writer isn't running, it fetches from infrastructure on every call.
func (s *Service) GetCurrentQuotaUsage(ctx context.Context, ownerID uuid.UUID, quotaEpoch time.Time) pb.QuotaUsage {
	// the beginning of the current quota period
	currentQuotaTime := s.GetQuotaPeriod(quotaEpoch).Floor(time.Now())

	// if not running, we fetch without caching
	if !s.running {
		usage, err := s.engine.Usage.ReadUsageSingle(ctx, ownerID, driver.UsageLabelQuotaMonth, currentQuotaTime)
		if err != nil {
			panic(err)
		}
		return usage
	}

	// create cache key
	timeBytes := bytesutil.IntToBytes(currentQuotaTime.Unix())
	cacheKey := string(append(ownerID.Bytes(), timeBytes...))

	// first, check cache for usage else get usage from store
	var usage pb.QuotaUsage
	val, err := s.usageCache.Get(cacheKey)
	if err == nil {
		// use cached value
		usage = val.(pb.QuotaUsage)
	} else if err != gcache.KeyNotFoundError {
		// unexpected
		panic(err)
	} else {
		// load from store
		usage, err = s.engine.Usage.ReadUsageSingle(ctx, ownerID, driver.UsageLabelQuotaMonth, currentQuotaTime)
		if err != nil {
			panic(err)
		}

		// write to cache
		s.usageCache.Set(cacheKey, usage)
	}

	// add buffer to quota
	s.mu.RLock()
	usageBuf := s.usageBuffer[ownerID]
	s.mu.RUnlock()
	usage.ReadOps += usageBuf.ReadOps
	usage.ReadRecords += usageBuf.ReadRecords
	usage.ReadBytes += usageBuf.ReadBytes
	usage.WriteOps += usageBuf.WriteOps
	usage.WriteRecords += usageBuf.WriteRecords
	usage.WriteBytes += usageBuf.WriteBytes
	usage.ScanOps += usageBuf.ScanOps
	usage.ScanBytes += usageBuf.ScanBytes

	return usage
}

// CheckReadQuota checks that the secret is within its quotas to trigger a read
func (s *Service) CheckReadQuota(ctx context.Context, secret models.Secret) error {
	if secret.IsAnonymous() {
		return nil
	}

	brq := secret.GetBillingReadQuota()
	if brq != nil {
		usage := s.GetCurrentQuotaUsage(ctx, secret.GetBillingOrganizationID(), secret.GetBillingQuotaEpoch())
		if usage.ReadBytes >= *brq {
			return fmt.Errorf("your organization has exhausted its monthly read quota")
		}
	}

	orq := secret.GetOwnerReadQuota()
	if orq != nil {
		usage := s.GetCurrentQuotaUsage(ctx, secret.GetOwnerID(), secret.GetOwnerQuotaEpoch())
		if usage.ReadBytes >= *orq {
			return fmt.Errorf("you have exhausted your monthly read quota")
		}
	}

	return nil
}

// CheckWriteQuota checks that the secret is within its quotas to trigger a write
func (s *Service) CheckWriteQuota(ctx context.Context, secret models.Secret) error {
	if secret.IsAnonymous() {
		return nil
	}

	bwq := secret.GetBillingWriteQuota()
	if bwq != nil {
		usage := s.GetCurrentQuotaUsage(ctx, secret.GetBillingOrganizationID(), secret.GetBillingQuotaEpoch())
		if usage.WriteBytes >= *bwq {
			return fmt.Errorf("your organization has exhausted its monthly write quota")
		}
	}

	owq := secret.GetOwnerWriteQuota()
	if owq != nil {
		usage := s.GetCurrentQuotaUsage(ctx, secret.GetOwnerID(), secret.GetOwnerQuotaEpoch())
		if usage.WriteBytes >= *owq {
			return fmt.Errorf("you have exhausted your monthly write quota")
		}
	}

	return nil
}

// CheckScanQuota checks that secret is within its quotas to trigger a warehouse query
func (s *Service) CheckScanQuota(ctx context.Context, secret models.Secret, estimatedScanBytes int64) error {
	if secret.IsAnonymous() {
		return fmt.Errorf("anonymous users cannot run warehouse queries")
	}

	bsq := secret.GetBillingScanQuota()
	if bsq != nil {
		usage := s.GetCurrentQuotaUsage(ctx, secret.GetBillingOrganizationID(), secret.GetBillingQuotaEpoch())
		if usage.ScanBytes >= *bsq {
			return fmt.Errorf("your organization has exhausted its monthly warehouse scan quota")
		} else if usage.ScanBytes+estimatedScanBytes > *bsq {
			return fmt.Errorf("your organization doesn't have a sufficient remaining scan quota to execute query (estimated at %d scanned bytes)", estimatedScanBytes)
		}
	}

	osq := secret.GetOwnerScanQuota()
	if osq != nil {
		usage := s.GetCurrentQuotaUsage(ctx, secret.GetOwnerID(), secret.GetOwnerQuotaEpoch())
		if usage.ScanBytes >= *osq {
			return fmt.Errorf("you have exhausted your monthly warehouse scan quota")
		} else if usage.ScanBytes+estimatedScanBytes > *osq {
			return fmt.Errorf("your don't have a sufficient remaining scan quota to execute query (estimated at %d scanned bytes)", estimatedScanBytes)
		}
	}

	return nil
}
