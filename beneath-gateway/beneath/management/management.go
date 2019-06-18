package management

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

// Init connects to management databases
func Init(postgresURL string, redisURL string) {
	DB = newDatabase(postgresURL)
	Redis = newRedis(redisURL)
}
