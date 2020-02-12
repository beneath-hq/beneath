package taskqueue

import (
	"context"

	"github.com/beneath-core/internal/hub"
)

// Submit queues a task for proocessing
func Submit(ctx context.Context, t Task) error {
	qt, err := EncodeTask(t)
	if err != nil {
		return err
	}

	return hub.Engine.QueueTask(ctx, qt)
}
