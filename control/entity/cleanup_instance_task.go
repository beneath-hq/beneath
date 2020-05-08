package entity

import (
	"context"
	"time"

	"gitlab.com/beneath-hq/beneath/control/taskqueue"
	"gitlab.com/beneath-hq/beneath/internal/hub"
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
	time.Sleep(getStreamCache().cacheLRUTime())

	err := hub.Engine.RemoveInstance(ctx, t.CachedStream, t.CachedStream, t.CachedStream)
	if err != nil {
		return err
	}

	err = hub.Engine.ClearUsage(ctx, t.CachedStream.InstanceID)
	if err != nil {
		return err
	}

	return nil
}
