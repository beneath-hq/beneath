package stream

import (
	"github.com/go-redis/redis/v7"

	"github.com/beneath-hq/beneath/bus"
	"github.com/beneath-hq/beneath/infra/db"
	"github.com/beneath-hq/beneath/infra/engine"
)

// Service has functionality for managing user and service secrets, including authentication
type Service struct {
	Bus    *bus.Bus
	DB     db.DB
	Redis  *redis.Client
	Engine *engine.Engine

	nameCache     *nameCache
	instanceCache *instanceCache
}

// New creates a new user service
func New(bus *bus.Bus, db db.DB, redis *redis.Client, engine *engine.Engine) *Service {
	s := &Service{
		Bus:    bus,
		DB:     db,
		Redis:  redis,
		Engine: engine,
	}
	s.initNameCache()
	s.initInstanceCache()
	s.Bus.AddSyncListener(s.organizationUpdated)
	s.Bus.AddSyncListener(s.projectUpdated)
	s.Bus.AddAsyncListener(s.streamUpdated)
	s.Bus.AddAsyncOrderedListener(s.streamDeleted)
	s.Bus.AddSyncListener(s.streamInstanceCreated)
	s.Bus.AddAsyncListener(s.streamInstanceUpdated)
	s.Bus.AddAsyncOrderedListener(s.streamInstanceDeleted)
	return s
}
