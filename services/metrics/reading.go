package metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/bluele/gcache"
	uuid "github.com/satori/go.uuid"

	pb "gitlab.com/beneath-hq/beneath/infrastructure/engine/proto"
	"gitlab.com/beneath-hq/beneath/pkg/bytesutil"
	"gitlab.com/beneath-hq/beneath/pkg/timeutil"
)

const (
	maxPeriods = 175 // roughly one week on an hourly basis
)

// GetHistoricalUsage returns usage info for the given length of time
func (b *Broker) GetHistoricalUsage(ctx context.Context, entityID uuid.UUID, period timeutil.Period, from time.Time, until time.Time) ([]time.Time, []pb.QuotaUsage, error) {
	// check is supported period
	if period != timeutil.PeriodHour && period != timeutil.PeriodMonth {
		return nil, nil, fmt.Errorf("usage is calculated only in hourly and monthly periods")
	}

	// if "until" is 0, set it to the current time
	if until.IsZero() {
		until = timeutil.Floor(time.Now(), period)
	}

	// check that the provided time range corresponds to keys in the table
	if from != timeutil.Floor(from, period) || until != timeutil.Floor(until, period) {
		return nil, nil, fmt.Errorf("from and until must exactly match period beginnings")
	}

	// prevent the retrieval of too many rows at once
	if period.Count(from, until) > maxPeriods {
		return nil, nil, fmt.Errorf("time span too long")
	}

	// read usage table and collect usage metrics
	var times []time.Time
	var usages []pb.QuotaUsage
	err := b.engine.ReadUsage(ctx, entityID, period, from, until, func(ts time.Time, usage pb.QuotaUsage) error {
		times = append(times, ts)
		usages = append(usages, usage)
		return nil
	})
	if err != nil {
		panic(fmt.Errorf("error reading from metrics table: %s", err.Error()))
	}

	return times, usages, nil
}

// GetCurrentUsage returns an ID's usage for the current monthly or hourly period
func (b *Broker) GetCurrentUsage(ctx context.Context, entityID uuid.UUID) pb.QuotaUsage {
	if !b.running {
		return b.loadCurrentUsage(ctx, entityID)
	}

	// create cache key
	timeBytes := bytesutil.IntToBytes(timeutil.Floor(time.Now(), timeutil.PeriodMonth).Unix())
	cacheKey := string(append(entityID.Bytes(), timeBytes...))

	// first, check cache for usage else get usage from store
	var usage pb.QuotaUsage
	val, err := b.usageCache.Get(cacheKey)
	if err == nil {
		// use cached value
		usage = val.(pb.QuotaUsage)
	} else if err != gcache.KeyNotFoundError {
		// unexpected
		panic(err)
	} else {
		// load from store
		usage = b.loadCurrentUsage(ctx, entityID)

		// write back to cache
		b.usageCache.Set(cacheKey, usage)
	}

	// add buffer to quota
	b.mu.RLock()
	usageBuf := b.buffer[entityID]
	b.mu.RUnlock()
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

// LoadCurrentUsage returns an ID's usage for the current monthly period
func (b *Broker) loadCurrentUsage(ctx context.Context, entityID uuid.UUID) pb.QuotaUsage {
	// load from bigtable
	usage, err := b.engine.ReadSingleUsage(ctx, entityID, timeutil.PeriodMonth, time.Now())
	if err != nil {
		panic(fmt.Errorf("error reading metrics: %s", err.Error()))
	}

	return usage
}
