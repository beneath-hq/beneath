package bi

import (
	"context"

	"gitlab.com/beneath-hq/beneath/bus"
	"gitlab.com/beneath-hq/beneath/infrastructure/db"
	"gitlab.com/beneath-hq/beneath/models"
)

// Service wraps business intelligence-related functionality
type Service struct {
	Bus *bus.Bus
	DB  db.DB
}

// New creates a new user service
func New(bus *bus.Bus, db db.DB) *Service {
	s := &Service{
		Bus: bus,
		DB:  db,
	}
	bus.AddSyncListener(s.UserCreated)
	return s
}

// UserCreated handles models.UserCreatedEvent events
func (s *Service) UserCreated(ctx context.Context, msg *models.UserCreatedEvent) error {
	err := s.Bus.PublishControlEvent(ctx, "user_create", map[string]interface{}{
		"organization_id": msg.User.BillingOrganizationID,
		"user_id":         msg.User.UserID,
	})
	if err != nil {
		return err
	}
	return nil
}
