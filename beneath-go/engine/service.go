package engine

import (
	"context"

	"golang.org/x/sync/errgroup"

	"github.com/beneath-core/beneath-go/engine/driver"
)

// RegisterProject is called when a project is created *or updated*
func (e *Engine) RegisterProject(ctx context.Context, p driver.Project) error {
	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		return e.Log.RegisterProject(ctx, p)
	})

	group.Go(func() error {
		return e.Lookup.RegisterProject(ctx, p)
	})

	group.Go(func() error {
		return e.Warehouse.RegisterProject(ctx, p)
	})

	return group.Wait()
}

// RemoveProject is called when a project is deleted
func (e *Engine) RemoveProject(ctx context.Context, p driver.Project) error {
	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		return e.Log.RemoveProject(ctx, p)
	})

	group.Go(func() error {
		return e.Lookup.RemoveProject(ctx, p)
	})

	group.Go(func() error {
		return e.Warehouse.RemoveProject(ctx, p)
	})

	return group.Wait()
}

// RegisterInstance is called when a new instance is created
func (e *Engine) RegisterInstance(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance) error {
	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		return e.Log.RegisterInstance(ctx, p, s, i)
	})

	group.Go(func() error {
		return e.Lookup.RegisterInstance(ctx, p, s, i)
	})

	group.Go(func() error {
		return e.Warehouse.RegisterInstance(ctx, p, s, i)
	})

	return group.Wait()
}

// PromoteInstance is called when an instance is promoted to be the main instance of the stream
func (e *Engine) PromoteInstance(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance) error {
	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		return e.Log.PromoteInstance(ctx, p, s, i)
	})

	group.Go(func() error {
		return e.Lookup.PromoteInstance(ctx, p, s, i)
	})

	group.Go(func() error {
		return e.Warehouse.PromoteInstance(ctx, p, s, i)
	})

	return group.Wait()
}

// RemoveInstance is called when an instance is deleted
func (e *Engine) RemoveInstance(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance) error {
	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		return e.Log.RemoveInstance(ctx, p, s, i)
	})

	group.Go(func() error {
		return e.Lookup.RemoveInstance(ctx, p, s, i)
	})

	group.Go(func() error {
		return e.Warehouse.RemoveInstance(ctx, p, s, i)
	})

	return group.Wait()
}
