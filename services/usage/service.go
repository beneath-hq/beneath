package usage

import (
	"sync"
	"time"

	"github.com/bluele/gcache"
	uuid "github.com/satori/go.uuid"

	"gitlab.com/beneath-hq/beneath/infrastructure/engine"
	pb "gitlab.com/beneath-hq/beneath/infrastructure/engine/proto"
	"gitlab.com/beneath-hq/beneath/pkg/timeutil"
)

// OpType defines a read or write operation
type OpType int

// OpType enum definition
const (
	OpTypeRead OpType = iota
	OpTypeWrite
	OpTypeScan
)

// Options for creating the usage service
type Options struct {
	CacheSize      int
	CommitInterval time.Duration
}

// Service reads and writes usage. For writes, it buffers updates and flushes them in the background.
type Service struct {
	opts   *Options
	engine *engine.Engine

	running          bool
	usageBuffer      map[uuid.UUID]pb.QuotaUsage // accumulates usage to flush
	quotaEpochBuffer map[uuid.UUID]time.Time     // stores quota epochs for IDs in usageBuffer
	mu               sync.RWMutex                // for buffers
	commitTicker     *time.Ticker                // periodically triggers a flush to BigTable

	usageCache gcache.Cache
}

// New initializes the service
func New(opts *Options, e *engine.Engine) *Service {
	b := &Service{
		opts:   opts,
		engine: e,
	}
	return b
}

// QuotaMonthDuration sets a "quota month" to a fixed-size 31 days
const QuotaMonthDuration = 31 * 24 * time.Hour

// GetQuotaPeriod gets the period to use for calculating quota-related timestamps for an owner
// with the given quota epoch.
func (s *Service) GetQuotaPeriod(quotaEpoch time.Time) timeutil.FixedOffsetPeriod {
	return timeutil.NewFixedOffsetPeriod(quotaEpoch, QuotaMonthDuration)
}
