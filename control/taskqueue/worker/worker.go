package worker

import (
	"context"
	"time"

	"gitlab.com/beneath-hq/beneath/control/taskqueue"
	pb "gitlab.com/beneath-hq/beneath/engine/proto"
	"gitlab.com/beneath-hq/beneath/hub"
	"gitlab.com/beneath-hq/beneath/pkg/log"
	"gitlab.com/beneath-hq/beneath/pkg/timeutil"
)

// Work runs a worker
func Work(ctx context.Context) error {
	log.S.Info("taskqueue worker started")
	return hub.Engine.ReadTasks(ctx, processTask)
}

// processWriteRequest is called (approximately once) for each task
func processTask(ctx context.Context, t *pb.QueuedTask) error {
	// metrics to track
	startTime := time.Now()

	// decode and run
	task, err := taskqueue.DecodeTask(t)
	if err == nil {
		err = task.Run(ctx)
	}

	// log
	log.S.Infow(
		"task processed",
		"uuid", t.UniqueId,
		"time", timeutil.FromUnixMilli(t.Timestamp),
		"name", t.Name,
		"err", err,
		"elapsed", time.Since(startTime),
	)

	// done
	return err
}
