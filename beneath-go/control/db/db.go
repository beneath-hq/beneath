package db

import (
	"github.com/beneath-core/beneath-go/engine"
	"github.com/go-pg/pg"
	"github.com/go-redis/redis"
)

var (
	// DB is the postgres connection
	DB *pg.DB

	// Redis connection
	Redis *redis.Client

	// Engine is the data plane
	Engine *engine.Engine
)

// InitPostgres sets up Postgres connection
func InitPostgres(postgresURL string) {
	DB = newDatabase(postgresURL)
}

// InitRedis sets up Redis connection
func InitRedis(redisURL string) {
	Redis = newRedis(redisURL)
}

// InitEngine sets up the engine connection
func InitEngine(streamsDriver string, tablesDriver string, warehouseDriver string) {
	Engine = engine.NewEngine(streamsDriver, tablesDriver, warehouseDriver)
}
