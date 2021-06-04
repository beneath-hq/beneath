package bi

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/beneath-hq/beneath/bus"
	"github.com/beneath-hq/beneath/infra/db"
	"github.com/beneath-hq/beneath/infra/mq"
	"github.com/beneath-hq/beneath/models"
	uuid "github.com/satori/go.uuid"
)

// Service wraps business intelligence-related functionality
type Service struct {
	Bus *bus.Bus
	DB  db.DB
	MQ  mq.MessageQueue
}

// New creates a new business intelligence service
func New(bus *bus.Bus, db db.DB, mq mq.MessageQueue) (*Service, error) {
	s := &Service{
		Bus: bus,
		DB:  db,
		MQ:  mq,
	}

	err := mq.RegisterTopic(controlEventsTopic, false)
	if err != nil {
		return nil, err
	}

	bus.AddSyncListener(s.UserCreated)

	return s, nil
}

// UserCreated handles models.UserCreatedEvent events
func (s *Service) UserCreated(ctx context.Context, msg *models.UserCreatedEvent) error {
	err := s.PublishControlEvent(ctx, "user_create", map[string]interface{}{
		"organization_id": msg.User.BillingOrganizationID,
		"user_id":         msg.User.UserID,
	})
	if err != nil {
		return err
	}
	return nil
}

// TODO: Remove this, but not right now

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
func (s *Service) PublishControlEvent(ctx context.Context, name string, data map[string]interface{}) error {
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
	if len(json) > s.MQ.MaxMessageSize() {
		return fmt.Errorf("control event message %v has invalid size", json)
	}
	return s.MQ.Publish(ctx, controlEventsTopic, json, nil)
}
