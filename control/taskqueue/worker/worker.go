package worker

import (
	"context"
	"time"

	"github.com/beneath-core/control/taskqueue"
	"github.com/beneath-core/internal/hub"
	pb "github.com/beneath-core/engine/proto"
	"github.com/beneath-core/pkg/log"
	"github.com/beneath-core/pkg/timeutil"
)

// Work runs a worker
func Work() error {
	log.S.Info("taskqueue worker started")
	return hub.Engine.ReadTasks(processTask)
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
