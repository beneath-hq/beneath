package redis

import (
	"github.com/go-redis/redis/v7"
)

// Options for opening a new Redis connection
type Options struct {
	URL string
}

// NewRedis opens a new Redis connection
func NewRedis(opts *Options) *redis.Client {
	redisOpts, err := redis.ParseURL(opts.URL)
	if err != nil {
		panic(err)
	}

	client := redis.NewClient(redisOpts)
	return client
}
