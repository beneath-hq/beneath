package metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/bluele/gcache"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/sync/semaphore"

	pb "gitlab.com/beneath-hq/beneath/infrastructure/engine/proto"
	"gitlab.com/beneath-hq/beneath/pkg/log"
	"gitlab.com/beneath-hq/beneath/pkg/timeutil"
)

// Run sets the broker to periodically commit buffered metrics until ctx is cancelled or an error occurs
func (b *Broker) Run(ctx context.Context) error {
	if b.running {
		panic(fmt.Errorf("Cannot call RunForever twice"))
	}

	b.running = true
	b.buffer = make(map[uuid.UUID]pb.QuotaUsage, b.opts.CacheSize)
	b.commitTicker = time.NewTicker(b.opts.CommitInterval)
	b.usageCache = gcache.New(b.opts.CacheSize).LRU().Build()

	for {
		select {
		case <-b.commitTicker.C:
			b.commitToTable()
		case <-ctx.Done():
			b.commitTicker.Stop()
			b.commitToTable()
			b.running = false
			log.S.Infow("metrics committed before stopping")
			return nil
		}
	}
}

var (
	maxWorkers = 100
	sem        = semaphore.NewWeighted(int64(maxWorkers))
)

// commitToTable commits a batch of accumulated metrics to BigTable every X seconds
func (b *Broker) commitToTable() error {
	ctx := context.Background()

	b.mu.Lock()
	buf := b.buffer
	b.buffer = make(map[uuid.UUID]pb.QuotaUsage, b.opts.CacheSize)
	b.mu.Unlock()

	// skip if nothing to upload
	if len(buf) == 0 {
		return nil
	}

	ts := time.Now()

	for id, usage := range buf {
		// when maxWorkers goroutines are in flight, Acquire blocks until one of the workers finishes.
		if err := sem.Acquire(ctx, 1); err != nil {
			log.S.Errorf("Failed to acquire semaphore: %v", err)
			break
		}

		go func(id uuid.UUID, usage pb.QuotaUsage) error {
			defer sem.Release(1)

			// commit metrics to monthly count
			err := b.engine.CommitUsage(ctx, id, timeutil.PeriodMonth, ts, usage)
			if err != nil {
				return err
			}

			// commit metrics to hourly count
			err = b.engine.CommitUsage(ctx, id, timeutil.PeriodHour, ts, usage)
			if err != nil {
				return err
			}

			return err
		}(id, usage)
	}

	// acquire all of the tokens to wait for any remaining workers to finish.
	if err := sem.Acquire(ctx, int64(maxWorkers)); err != nil {
		log.S.Errorf("Failed to acquire semaphore: %v", err)
	}

	// release all the tokens to be ready for the next batch
	sem.Release(int64(maxWorkers))

	// reset usage cache
	b.usageCache.Purge()

	// log
	elapsed := time.Since(ts)
	log.S.Infow(
		"metrics write",
		"ids", len(buf),
		"elapsed", elapsed,
	)

	// done
	return nil
}

// TrackRead records metrics for reads
func (b *Broker) TrackRead(id uuid.UUID, nrecords int64, nbytes int64) {
	b.trackOp(OpTypeRead, id, nrecords, nbytes)
}

// TrackWrite records metrics for writes
func (b *Broker) TrackWrite(id uuid.UUID, nrecords int64, nbytes int64) {
	b.trackOp(OpTypeWrite, id, nrecords, nbytes)
}

// TrackScan records metrics for query scans
func (b *Broker) TrackScan(id uuid.UUID, nbytes int64) {
	b.trackOp(OpTypeScan, id, 0, nbytes)
}

// trackOp records the metrics for a given ID (project, user, or secret)
func (b *Broker) trackOp(op OpType, id uuid.UUID, nrecords int64, nbytes int64) {
	b.mu.Lock()
	u := b.buffer[id]
	if op == OpTypeRead {
		u.ReadOps++
		u.ReadRecords += nrecords
		u.ReadBytes += nbytes
	} else if op == OpTypeWrite {
		u.WriteOps++
		u.WriteRecords += nrecords
		u.WriteBytes += nbytes
	} else if op == OpTypeScan {
		u.ScanOps++
		u.ScanBytes += nbytes
	} else {
		b.mu.Unlock()
		panic(fmt.Errorf("unrecognized op type '%d'", op))
	}
	b.buffer[id] = u
	b.mu.Unlock()
}
