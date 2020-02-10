package engine

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/proto"

	"github.com/beneath-core/beneath-go/core/ctxutil"
	pb "github.com/beneath-core/beneath-go/engine/proto"
)

const (
	tasksTopic        = "tasks"
	tasksSubscription = "tasks-worker"
)

// QueueTask queues a task for processing
func (e *Engine) QueueTask(ctx context.Context, t *pb.QueuedTask) error {
	msg, err := proto.Marshal(t)
	if err != nil {
		panic(err)
	}
	if len(msg) > e.MQ.MaxMessageSize() {
		return fmt.Errorf("task %v has invalid size", t)
	}
	return e.MQ.Publish(ctx, tasksTopic, msg)
}

// ReadTasks reads queued tasks
func (e *Engine) ReadTasks(fn func(context.Context, *pb.QueuedTask) error) error {
	ctx := ctxutil.WithCancelOnTerminate(context.Background())
	return e.MQ.Subscribe(ctx, tasksTopic, tasksSubscription, true, func(ctx context.Context, msg []byte) error {
		t := &pb.QueuedTask{}
		err := proto.Unmarshal(msg, t)
		if err != nil {
			return err
		}
		return fn(ctx, t)
	})
}
