package entity

import (
	"context"
	"time"

	uuid "github.com/satori/go.uuid"

	"gitlab.com/beneath-hq/beneath/control/taskqueue"
	"gitlab.com/beneath-hq/beneath/hub"
)

// CleanupInstanceTask is a task that removes all data and tables related to an instance
type CleanupInstanceTask struct {
	InstanceID   uuid.UUID
	CachedStream *CachedStream
}

// register task
func init() {
	taskqueue.RegisterTask(&CleanupInstanceTask{})
}

// Run triggers the task
func (t *CleanupInstanceTask) Run(ctx context.Context) error {
	time.Sleep(getStreamCache().cacheLRUTime())

	err := hub.Engine.RemoveInstance(ctx, t.CachedStream, t.CachedStream, EfficientStreamInstance(t.InstanceID))
	if err != nil {
		return err
	}

	err = hub.Engine.ClearUsage(ctx, t.InstanceID)
	if err != nil {
		return err
	}

	return nil
}
