package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	uuid "github.com/satori/go.uuid"
)

const (
	controlEventsTopic        = "control-events"
	controlEventsSubscription = "control-events-worker"
)

// ControlEvent describes a control-plane event
type ControlEvent struct {
	// TODO... check to see if we need these json tags. look at stripe library for their events
	ID        string `json:"unique_id"`
	Name      string `json:"name"`
	Timestamp int64  `json:"timestamp"`
	Data      []byte `json:"data"`
}

// PublishControlEvent publishes a control-plane event to controlEventsTopic
// for external consumption (e.g. for BI purposes)
func (e *Engine) PublishControlEvent(ctx context.Context, name string, data interface{}) error {
	id := uuid.NewV4()
	msg := map[string]interface{}{
		"id":   id,
		"name": name,
		"time": time.Now(),
		"data": data,
	}
	json, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	if len(msg) > e.MQ.MaxMessageSize() {
		return fmt.Errorf("control event message has invalid size: '%s'", json)
	}
	return e.MQ.Publish(ctx, controlEventsTopic, json)
}

// SubscribeControlEvents subscribes to all control events
// TODO: these are json encoded NOT protocol buffers
func (e *Engine) SubscribeControlEvents(ctx context.Context, fn func(context.Context, *ControlEvent) error) error {
	return e.MQ.Subscribe(ctx, controlEventsTopic, controlEventsSubscription, true, func(ctx context.Context, msg []byte) error {
		t := &ControlEvent{}
		err := json.Unmarshal(msg, t)
		if err != nil {
			return err
		}
		return fn(ctx, t)
	})
}
