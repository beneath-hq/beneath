package metrics

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/beneath-core/beneath-go/core/codec/ext/tuple"
	"github.com/beneath-core/beneath-go/db"
	pb "github.com/beneath-core/beneath-go/proto"
	uuid "github.com/satori/go.uuid"
)

// Broker coordinates the buffer and ticker
type Broker struct {
	// accumulates metrics to commit
	buffer map[uuid.UUID]pb.QuotaUsage

	// periodically triggers a commit to BigTable
	commitTicker *time.Ticker

	// Lock
	mu sync.RWMutex
}

// OpType defines a read or write operation
type OpType int

// OpType enum definition
const (
	OpTypeRead OpType = iota
	OpTypeWrite
)

const (
	commitInterval = 5 * time.Second

	// MonthlyPeriod represents the duration for which usage is monitored
	MonthlyPeriod = "M"

	// HourlyPeriod represents the hourly checkpoints for which usage is monitored
	HourlyPeriod = "H"

	// ReadUsage represents the metrics for read operations
	ReadUsage = "r"

	// WriteUsage represents the metrics for write operations
	WriteUsage = "w"
)

// NewBroker initializes the Broker
func NewBroker() *Broker {
	// create the Broker
	b := &Broker{
		buffer:       make(map[uuid.UUID]pb.QuotaUsage),
		commitTicker: time.NewTicker(commitInterval),
	}

	// start ticking for batch commits to BigTable
	go b.tick()

	// done
	return b
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

// TrackRead records metrics for reads
func (b *Broker) TrackRead(id uuid.UUID, nrecords int64, nbytes int64) {
	b.trackOp(OpTypeRead, id, nrecords, nbytes)
}

// TrackWrite records metrics for writes
func (b *Broker) TrackWrite(id uuid.UUID, nrecords int64, nbytes int64) {
	b.trackOp(OpTypeWrite, id, nrecords, nbytes)
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

// commitToTable commits a batch of accumulated metrics to BigTable every X seconds
func (b *Broker) commitToTable() error {
	ctx := context.Background()

	b.mu.Lock()
	buf := b.buffer
	b.buffer = make(map[uuid.UUID]pb.QuotaUsage)
	b.mu.Unlock()

	ts := time.Now()

	// TODO: use "golang.org/x/sync/semaphore" to implement a worker pool of goroutines to upload metrics in batch
	for id, usage := range buf {
		// commit metrics to monthly count
		rowKey := metricsKey(MonthlyPeriod, id, ts)

		err := db.Engine.Tables.CommitUsage(ctx, rowKey, usage)
		if err != nil {
			return err
		}

		// commit metrics to hourly count
		rowKey = metricsKey(HourlyPeriod, id, ts)
		err = db.Engine.Tables.CommitUsage(ctx, rowKey, usage)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetCurrentUsage returns an ID's usage for a given monthly/hourly period
func (b *Broker) GetCurrentUsage(ctx context.Context, id uuid.UUID, period string) pb.QuotaUsage {
	// create row filter
	keyPrefix := metricsKey(period, id, time.Now())

	usage := pb.QuotaUsage{}
	counter := 0

	// read table and collect metrics
	err := db.Engine.Tables.ReadUsage(ctx, keyPrefix, func(key []byte, u pb.QuotaUsage) error {
		if counter != 0 {
			return fmt.Errorf("the metrics lookup returned multiple keys")
		}
		usage = u
		counter++
		return nil
	})
	if err != nil {
		panic(fmt.Errorf("Error reading from Metrics table: %s", err.Error()))
	}

	return usage
}

// metricsKeyPrefix
func metricsKeyPrefix(period string, id uuid.UUID) []byte {
	return append([]byte(period), id.Bytes()...)
}

// metricsKey returns the binary representation of the key
func metricsKey(period string, id uuid.UUID, ts time.Time) []byte {
	// round ts to period
	ts = ts.UTC()
	switch period {
	case HourlyPeriod:
		ts = time.Date(ts.Year(), ts.Month(), ts.Day(), ts.Hour(), 0, 0, 0, time.UTC)
	case MonthlyPeriod:
		ts = time.Date(ts.Year(), ts.Month(), 1, 0, 0, 0, 0, time.UTC)
	default:
		panic(fmt.Errorf("unknown period '%s'", period))
	}

	// turn ts into binary representation
	unix := ts.Unix()
	timeEncoded := tuple.Tuple{unix}.Pack()

	// append to metrics key prefix
	return append(metricsKeyPrefix(period, id), timeEncoded...)
}
