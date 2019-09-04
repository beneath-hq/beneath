package db

import (
	"github.com/go-redis/redis"
)

func newRedis(redisURL string) *redis.Client {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		panic(err)
	}

	client := redis.NewClient(opts)
	return client
}
