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
	ID        uuid.UUID              `json:"id"`
	Name      string                 `json:"name"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// PublishControlEvent publishes a control-plane event to controlEventsTopic
// for external consumption (e.g. for BI purposes)
func (e *Engine) PublishControlEvent(ctx context.Context, name string, data map[string]interface{}) error {
	msg := ControlEvent{
		ID:        uuid.NewV4(),
		Name:      name,
		Timestamp: time.Now(),
		Data:      data,
	}
	json, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	if len(json) > e.MQ.MaxMessageSize() {
		return fmt.Errorf("control event message %v has invalid size", json)
	}
	return e.MQ.Publish(ctx, controlEventsTopic, json)
}

// SubscribeControlEvents subscribes to all control events
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
