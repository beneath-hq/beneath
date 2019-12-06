package engine

import (
	"context"

	pb "github.com/beneath-core/beneath-go/proto"
)

// QueueTask queues a task for processing
func (e *Engine) QueueTask(ctx context.Context, t *pb.QueuedTask) error {
	panic("todo")
}

// ReadTasks reads queued tasks
func (e *Engine) ReadTasks(fn func(context.Context, *pb.QueuedTask) error) error {
	panic("todo")
}
