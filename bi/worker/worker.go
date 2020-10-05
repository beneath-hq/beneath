package worker

import (
	"context"
	"time"

	"gitlab.com/beneath-hq/beneath/bi/events"
	"gitlab.com/beneath-hq/beneath/engine"
	"gitlab.com/beneath-hq/beneath/internal/hub"
	"gitlab.com/beneath-hq/beneath/pkg/log"
)

// Work runs a worker
func Work(ctx context.Context) error {
	log.S.Info("business intelligence worker started")
	return hub.Engine.SubscribeControlEvents(ctx, processControlEvent)
}

// processControlEvent is called (approximately once) for each task
func processControlEvent(ctx context.Context, event *engine.ControlEvent) error {
	// metrics to track
	startTime := time.Now()
	var err error

	// route to event to processing function
	switch event.Name {
	case "user_create":
		err = events.HandleUserCreate(ctx, event)
	default:
		log.S.Info("unknown control event")
	}

	// log
	log.S.Infow(
		"control event processed",
		"uuid", event.ID,
		"time", event.Timestamp.Unix(),
		"name", event.Name,
		"err", err,
		"elapsed", time.Since(startTime),
	)

	// done
	return err
}
