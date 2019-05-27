package beneath

import (
	"log"

	"github.com/go-redis/redis"
)

var (
	// Redis connection
	Redis *redis.Client
)

func init() {
	opts, err := redis.ParseURL(Config.RedisURL)
	if err != nil {
		log.Fatalf("redis: %s", err.Error())
	}

	Redis = redis.NewClient(opts)
}
