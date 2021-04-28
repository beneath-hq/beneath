package usage

import (
	"context"
	"fmt"
	"time"

	"github.com/bluele/gcache"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/sync/semaphore"

	"github.com/beneath-hq/beneath/infra/engine/driver"
	pb "github.com/beneath-hq/beneath/infra/engine/proto"
	"github.com/beneath-hq/beneath/pkg/timeutil"
)

var (
	maxWorkers = 100
	sem        = semaphore.NewWeighted(int64(maxWorkers))
)

// Run periodically flushes buffered usage data until ctx is cancelled or an error occurs
func (s *Service) Run(ctx context.Context) error {
	if s.running {
		panic(fmt.Errorf("Cannot call RunForever twice"))
	}

	s.running = true
	s.usageBuffer = make(map[uuid.UUID]pb.QuotaUsage, s.opts.CacheSize)
	s.quotaEpochBuffer = make(map[uuid.UUID]time.Time, s.opts.CacheSize)
	s.commitTicker = time.NewTicker(s.opts.CommitInterval)
	s.usageCache = gcache.New(s.opts.CacheSize).LRU().Build()

	for {
		select {
		case <-s.commitTicker.C:
			s.commitToTable()
		case <-ctx.Done():
			s.commitTicker.Stop()
			s.commitToTable()
			s.running = false
			s.logger.Infow("buffered usage flushed before stopping")
			return nil
		}
	}
}

// commitToTable flushes a batch of accumulated usage data to the engine (called by Run every X seconds)
func (s *Service) commitToTable() error {
	ctx := context.Background()

	s.mu.Lock()
	usageBuffer := s.usageBuffer
	quotaEpochBuffer := s.quotaEpochBuffer
	s.usageBuffer = make(map[uuid.UUID]pb.QuotaUsage, s.opts.CacheSize)
	s.quotaEpochBuffer = make(map[uuid.UUID]time.Time, s.opts.CacheSize)
	s.mu.Unlock()

	// skip if nothing to upload
	if len(usageBuffer) == 0 {
		return nil
	}

	now := time.Now()

	for id, usage := range usageBuffer {
		// when maxWorkers goroutines are in flight, Acquire blocks until one of the workers finishes.
		if err := sem.Acquire(ctx, 1); err != nil {
			s.logger.Errorf("failed to acquire semaphore: %v", err)
			break
		}

		id := id
		usage := usage
		go func(id uuid.UUID, usage pb.QuotaUsage) error {
			defer sem.Release(1)

			// add usage to monthly count
			err := s.engine.Usage.WriteUsage(ctx, id, driver.UsageLabelMonthly, timeutil.Floor(now, timeutil.PeriodMonth), usage)
			if err != nil {
				return err
			}

			// add usage to hourly count
			err = s.engine.Usage.WriteUsage(ctx, id, driver.UsageLabelHourly, timeutil.Floor(now, timeutil.PeriodHour), usage)
			if err != nil {
				return err
			}

			// add usage to quota month count (if applicable)
			if quotaEpoch, ok := quotaEpochBuffer[id]; ok {
				err = s.engine.Usage.WriteUsage(ctx, id, driver.UsageLabelQuotaMonth, s.GetQuotaPeriod(quotaEpoch).Floor(now), usage)
				if err != nil {
					return err
				}
			}

			return nil
		}(id, usage)
	}

	// acquire all of the tokens to wait for any remaining workers to finish.
	if err := sem.Acquire(ctx, int64(maxWorkers)); err != nil {
		s.logger.Errorf("failed to acquire semaphore: %v", err)
	}

	// release all the tokens to be ready for the next batch
	sem.Release(int64(maxWorkers))

	// reset usage cache
	s.usageCache.Purge()

	// log
	elapsed := time.Since(now)
	s.logger.Infow(
		"usage flush",
		"ids", len(usageBuffer),
		"quota_ids", len(quotaEpochBuffer),
		"elapsed", elapsed,
	)

	// done
	return nil
}
