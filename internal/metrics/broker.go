package metrics

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bluele/gcache"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/sync/semaphore"

	pb "gitlab.com/beneath-org/beneath/engine/proto"
	"gitlab.com/beneath-org/beneath/internal/hub"
	"gitlab.com/beneath-org/beneath/pkg/bytesutil"
	"gitlab.com/beneath-org/beneath/pkg/log"
	"gitlab.com/beneath-org/beneath/pkg/timeutil"
)

// Broker coordinates the buffer and ticker
type Broker struct {
	// options set by NewBroker
	cacheSize      int
	commitInterval time.Duration

	// accumulates metrics to commit
	buffer map[uuid.UUID]pb.QuotaUsage

	// periodically triggers a commit to BigTable
	commitTicker *time.Ticker

	// Lock
	mu sync.RWMutex

	// usage cache
	usageCache gcache.Cache
}

// OpType defines a read or write operation
type OpType int

// OpType enum definition
const (
	OpTypeRead OpType = iota
	OpTypeWrite
)

// NewBroker initializes the Broker
func NewBroker(cacheSize int, commitInterval time.Duration) *Broker {
	// create the Broker
	b := &Broker{
		cacheSize:      cacheSize,
		commitInterval: commitInterval,
		buffer:         make(map[uuid.UUID]pb.QuotaUsage, cacheSize),
		commitTicker:   time.NewTicker(commitInterval),
		usageCache:     gcache.New(cacheSize).LRU().Build(),
	}

	// done
	return b
}

// RunForever sets the broker to periodically commit buffered metrics until ctx is cancelled or an error occurs
func (b *Broker) RunForever(ctx context.Context) {
	for {
		select {
		case <-b.commitTicker.C:
			b.commitToTable()
		case <-ctx.Done():
			b.commitTicker.Stop()
			b.commitToTable()
			log.S.Infow("metrics committed before stopping")
			return
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
	b.buffer = make(map[uuid.UUID]pb.QuotaUsage, b.cacheSize)
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
			err := hub.Engine.CommitUsage(ctx, id, timeutil.PeriodMonth, ts, usage)
			if err != nil {
				return err
			}

			// commit metrics to hourly count
			err = hub.Engine.CommitUsage(ctx, id, timeutil.PeriodHour, ts, usage)
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
	} else {
		b.mu.Unlock()
		panic(fmt.Errorf("unrecognized op type '%d'", op))
	}
	b.buffer[id] = u
	b.mu.Unlock()
}

// GetCurrentUsage returns an ID's usage for the current monthly or hourly period
func (b *Broker) GetCurrentUsage(ctx context.Context, entityID uuid.UUID) pb.QuotaUsage {
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
		usage = GetCurrentUsage(ctx, entityID)

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

	return usage
}
