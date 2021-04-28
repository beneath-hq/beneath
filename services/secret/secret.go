package secret

import (
	"context"

	"github.com/beneath-hq/beneath/bus"
	"github.com/beneath-hq/beneath/infra/db"
	"github.com/beneath-hq/beneath/models"
	"github.com/go-redis/redis/v7"
	uuid "github.com/satori/go.uuid"
)

const (
	// TokenFlagsService is used as flags byte for service secret tokens
	TokenFlagsService = byte(0x80)

	// TokenFlagsUser is used as flags byte for user secret tokens
	TokenFlagsUser = byte(0x81)
)

// Service has functionality for managing user and service secrets, including authentication
type Service struct {
	Bus   *bus.Bus
	DB    db.DB
	Redis *redis.Client

	cache *Cache
}

// New creates a new user service
func New(bus *bus.Bus, db db.DB, redis *redis.Client) *Service {
	s := &Service{
		Bus:   bus,
		DB:    db,
		Redis: redis,
	}
	s.initCache()
	s.Bus.AddSyncListener(s.userUpdated)
	s.Bus.AddSyncListener(s.organizationUpdated)
	s.Bus.AddSyncListener(s.serviceUpdated)
	s.Bus.AddSyncListener(s.serviceDeleted)
	return s
}

// FindUserSecret finds a secret
func (s *Service) FindUserSecret(ctx context.Context, secretID uuid.UUID) *models.UserSecret {
	secret := &models.UserSecret{UserSecretID: secretID}
	err := s.DB.GetDB(ctx).ModelContext(ctx, secret).WherePK().Select()
	if !db.AssertFoundOne(err) {
		return nil
	}
	return secret
}

// FindUserSecrets finds all the user's secrets
func (s *Service) FindUserSecrets(ctx context.Context, userID uuid.UUID) []*models.UserSecret {
	var secrets []*models.UserSecret
	err := s.DB.GetDB(ctx).ModelContext(ctx, &secrets).Where("user_id = ?", userID).Limit(1000).Order("created_on DESC").Select()
	if err != nil {
		panic(err)
	}
	return secrets
}

// FindServiceSecret finds a service secret
func (s *Service) FindServiceSecret(ctx context.Context, secretID uuid.UUID) *models.ServiceSecret {
	secret := &models.ServiceSecret{ServiceSecretID: secretID}
	err := s.DB.GetDB(ctx).ModelContext(ctx, secret).WherePK().Select()
	if !db.AssertFoundOne(err) {
		return nil
	}
	return secret
}

// FindServiceSecrets finds all the service's secrets
func (s *Service) FindServiceSecrets(ctx context.Context, serviceID uuid.UUID) []*models.ServiceSecret {
	var secrets []*models.ServiceSecret
	err := s.DB.GetDB(ctx).ModelContext(ctx, &secrets).Where("service_id = ?", serviceID).Limit(1000).Order("created_on DESC").Select()
	if err != nil {
		panic(err)
	}
	return secrets
}
