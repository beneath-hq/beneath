package engine

import (
	"context"

	"github.com/beneath-core/beneath-go/engine/driver"
)

// RegisterProject is called when a project is created *or updated*
func (e *Engine) RegisterProject(ctx context.Context, p driver.Project) error {
	panic("todo")
}

// RemoveProject is called when a project is deleted
func (e *Engine) RemoveProject(ctx context.Context, p driver.Project) error {
	panic("todo")
}

// RegisterInstance is called when a new instance is created
func (e *Engine) RegisterInstance(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance) error {
	panic("todo")
}

// PromoteInstance is called when an instance is promoted to be the main instance of the stream
func (e *Engine) PromoteInstance(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance) error {
	panic("todo")
}

// RemoveInstance is called when an instance is deleted
func (e *Engine) RemoveInstance(ctx context.Context, p driver.Project, s driver.Stream, i driver.StreamInstance) error {
	panic("todo")

	// TODO:
	// delete in bigtable records
	// err = db.Engine.Tables.ClearRecords(ctx, t.InstanceID)
	// if err != nil {
	// 	return err
	// }
}
