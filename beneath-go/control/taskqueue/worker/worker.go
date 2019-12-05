package worker

import (
	"context"
	"time"

	"github.com/beneath-core/beneath-go/control/taskqueue"
	"github.com/beneath-core/beneath-go/core/log"
	"github.com/beneath-core/beneath-go/core/timeutil"
	"github.com/beneath-core/beneath-go/db"
	pb "github.com/beneath-core/beneath-go/proto"
)

// Work runs a worker
func Work() error {
	log.S.Info("taskqueue worker started")
	return db.Engine.ReadTasks(processTask)
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
