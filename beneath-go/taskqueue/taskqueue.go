package taskqueue

import (
	"context"
	"time"

	"github.com/beneath-core/beneath-go/core/timeutil"

	"github.com/beneath-core/beneath-go/core/log"
	"github.com/beneath-core/beneath-go/db"
	pb "github.com/beneath-core/beneath-go/proto"
)

// Submit queues a task for proocessing
func Submit(ctx context.Context, t Task) error {
	qt, err := encodeTask(t)
	if err != nil {
		return err
	}

	return db.Engine.Streams.QueueTask(ctx, qt)
}

// Work runs a worker
func Work() error {
	log.S.Info("taskqueue worker started")
	return db.Engine.Streams.ReadTasks(processTask)
}

// processWriteRequest is called (approximately once) for each task
func processTask(ctx context.Context, t *pb.QueuedTask) error {
	// metrics to track
	startTime := time.Now()

	// decode and run
	task, err := decodeTask(t)
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
