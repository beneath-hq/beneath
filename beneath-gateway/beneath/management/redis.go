package management

import (
	"log"

	"github.com/go-redis/redis"
)

func newRedis(redisURL string) *redis.Client {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("redis: %s", err.Error())
	}

	client := redis.NewClient(opts)
	return client
}
