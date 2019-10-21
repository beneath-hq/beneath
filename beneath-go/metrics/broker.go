package metrics

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bluele/gcache"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/sync/semaphore"

	"github.com/beneath-core/beneath-go/core/log"
	"github.com/beneath-core/beneath-go/core/timeutil"
	"github.com/beneath-core/beneath-go/db"
	pb "github.com/beneath-core/beneath-go/proto"
)

// Broker coordinates the buffer and ticker
type Broker struct {
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

const (
	cacheSize = 2500

	commitInterval = 30 * time.Second

	// ReadUsage represents the metrics for read operations
	ReadUsage = "r"

	// WriteUsage represents the metrics for write operations
	WriteUsage = "w"
)

// NewBroker initializes the Broker
func NewBroker() *Broker {
	// create the Broker
	b := &Broker{
		buffer:       make(map[uuid.UUID]pb.QuotaUsage, cacheSize),
		commitTicker: time.NewTicker(commitInterval),
		usageCache:   gcache.New(cacheSize).LRU().Build(),
	}

	// start ticking for batch commits to BigTable
	go b.tick()

	// done
	return b
}

// tick continuously writes the buffer to BigTable every X seconds
func (b *Broker) tick() {
	for {
		select {
		case <-b.commitTicker.C:
			b.commitToTable()
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
	b.buffer = make(map[uuid.UUID]pb.QuotaUsage, cacheSize)
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
			rowKey := metricsKey(timeutil.PeriodMonth, id, ts)
			err := db.Engine.Tables.CommitUsage(ctx, rowKey, usage)
			if err != nil {
				return err
			}

			// commit metrics to hourly count
			rowKey = metricsKey(timeutil.PeriodHour, id, ts)
			err = db.Engine.Tables.CommitUsage(ctx, rowKey, usage)
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
	// create row filter
	key := metricsKey(timeutil.PeriodMonth, entityID, time.Now())

	// first, check cache for usage else get usage from bigtable
	var usage pb.QuotaUsage
	val, err := b.usageCache.Get(string(key))
	if err == nil {
		// use cached value
		usage = val.(pb.QuotaUsage)
	} else if err != gcache.KeyNotFoundError {
		// unexpected
		panic(err)
	} else {
		// load from bigtable
		usage := GetCurrentUsage(ctx, entityID)

		// write back to cache
		b.usageCache.Set(string(key), usage)
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
