package user

import (
	"context"

	uuid "github.com/satori/go.uuid"

	"github.com/beneath-hq/beneath/bus"
	"github.com/beneath-hq/beneath/infra/db"
	"github.com/beneath-hq/beneath/models"
)

// Service contains functionality for finding and creating users
type Service struct {
	Bus *bus.Bus
	DB  db.DB
}

// New creates a new user service
func New(bus *bus.Bus, db db.DB) *Service {
	return &Service{
		Bus: bus,
		DB:  db,
	}
}

// FindUser returns the matching user or nil
func (s *Service) FindUser(ctx context.Context, userID uuid.UUID) *models.User {
	user := &models.User{
		UserID: userID,
	}
	err := s.DB.GetDB(ctx).ModelContext(ctx, user).WherePK().Column("user.*").Select()
	if !db.AssertFoundOne(err) {
		return nil
	}
	return user
}
