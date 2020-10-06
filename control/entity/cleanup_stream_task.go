package entity

import (
	"context"

	uuid "github.com/satori/go.uuid"

	"gitlab.com/beneath-hq/beneath/control/taskqueue"
	"gitlab.com/beneath-hq/beneath/hub"
)

// CleanupStreamTask is a task triggered after a stream is deleted
type CleanupStreamTask struct {
	StreamID uuid.UUID
}

// register task
func init() {
	taskqueue.RegisterTask(&CleanupStreamTask{})
}

// Run triggers the task
func (t *CleanupStreamTask) Run(ctx context.Context) error {
	err := hub.Engine.ClearUsage(ctx, t.StreamID)
	if err != nil {
		return err
	}

	return nil
}
