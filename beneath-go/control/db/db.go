package db

import (
	"github.com/go-pg/pg"
	"github.com/go-redis/redis"
)

var (
	// DB is the postgres connection
	DB *pg.DB

	// Redis connection
	Redis *redis.Client
)

// InitPostgres sets up Postgres connection
func InitPostgres(postgresURL string) {
	DB = newDatabase(postgresURL)
}

// InitRedis sets up Redis connection
func InitRedis(redisURL string) {
	Redis = newRedis(redisURL)
}
