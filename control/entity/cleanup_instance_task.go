package entity

import (
	"context"

	"github.com/beneath-core/control/taskqueue"
	"github.com/beneath-core/db"
)

// CleanupInstanceTask is a task that removes all data and tables related to an instance
type CleanupInstanceTask struct {
	CachedStream *CachedStream
}

// register task
func init() {
	taskqueue.RegisterTask(&CleanupInstanceTask{})
}

// Run triggers the task
func (t *CleanupInstanceTask) Run(ctx context.Context) error {
	err := db.Engine.RemoveInstance(ctx, t.CachedStream, t.CachedStream, t.CachedStream)
	if err != nil {
		return err
	}

	return nil
}
