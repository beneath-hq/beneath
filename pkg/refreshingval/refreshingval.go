package refreshingval

import (
	"context"
	"sync"
	"time"
)

// Getter is a func that gets a refreshed value
type Getter func(context.Context) interface{}

// RefreshingValue caches a value and ensures its refreshed after a duration
type RefreshingValue struct {
	keep   time.Duration
	getter Getter

	mu        sync.Mutex
	value     interface{}
	refreshed time.Time
}

// New creates a new RefreshingValue
func New(keep time.Duration, getter Getter) *RefreshingValue {
	return &RefreshingValue{
		keep:   keep,
		getter: getter,
	}
}

// Get returns (and possibly refreshes) the wrapped value
func (r *RefreshingValue) Get(ctx context.Context) (res interface{}) {
	r.mu.Lock()
	if time.Since(r.refreshed) <= r.keep {
		res = r.value
		r.mu.Unlock()
		return res
	}

	r.value = r.getter(ctx)
	r.refreshed = time.Now()
	res = r.value
	r.mu.Unlock()

	return res
}
