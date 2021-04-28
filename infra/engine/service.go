package engine

import (
	"context"

	"golang.org/x/sync/errgroup"

	"github.com/beneath-hq/beneath/infra/engine/driver"
)

// RegisterInstance is called when a new instance is created
func (e *Engine) RegisterInstance(ctx context.Context, s driver.Stream, i driver.StreamInstance) error {
	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		return e.Lookup.RegisterInstance(ctx, s, i)
	})

	if s.GetUseWarehouse() {
		group.Go(func() error {
			return e.Warehouse.RegisterInstance(ctx, s, i)
		})
	}

	return group.Wait()
}

// RemoveInstance is called when an instance is deleted
func (e *Engine) RemoveInstance(ctx context.Context, s driver.Stream, i driver.StreamInstance) error {
	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		return e.Lookup.RemoveInstance(ctx, s, i)
	})

	if s.GetUseWarehouse() {
		group.Go(func() error {
			return e.Warehouse.RemoveInstance(ctx, s, i)
		})
	}

	return group.Wait()
}
