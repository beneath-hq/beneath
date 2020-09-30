package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	uuid "github.com/satori/go.uuid"
)

const (
	controlEventsTopic = "control-events"
)

// LogControlEvent publishes a control-plane event to controlEventsTopic
// for external consumption (e.g. for BI purposes)
func (e *Engine) LogControlEvent(ctx context.Context, name string, data interface{}) error {
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
