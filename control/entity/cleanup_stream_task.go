package entity

import (
	"context"

	uuid "github.com/satori/go.uuid"

	"github.com/beneath-core/control/taskqueue"
	"github.com/beneath-core/internal/hub"
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
