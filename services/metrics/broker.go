package metrics

import (
	"sync"
	"time"

	"github.com/bluele/gcache"
	uuid "github.com/satori/go.uuid"

	"gitlab.com/beneath-hq/beneath/infrastructure/engine"
	pb "gitlab.com/beneath-hq/beneath/infrastructure/engine/proto"
)

// Options for creating a metrics broker
type Options struct {
	CacheSize      int
	CommitInterval time.Duration
}

// Broker reads and writes metrics. For writes, it buffers updates and flushes them in the background.
type Broker struct {
	opts   *Options
	engine *engine.Engine

	running      bool
	buffer       map[uuid.UUID]pb.QuotaUsage // accumulates metrics to commit
	mu           sync.RWMutex                // for buffer
	commitTicker *time.Ticker                // periodically triggers a commit to BigTable

	usageCache gcache.Cache
}

// OpType defines a read or write operation
type OpType int

// OpType enum definition
const (
	OpTypeRead OpType = iota
	OpTypeWrite
	OpTypeScan
)

// New initializes the Broker
func New(opts *Options, e *engine.Engine) *Broker {
	// create the Broker
	b := &Broker{
		opts:   opts,
		engine: e,
	}

	// done
	return b
}
