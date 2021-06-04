package middleware

import (
	"github.com/beneath-hq/beneath/services/secret"
	"github.com/go-redis/redis/v7"
	"github.com/go-redis/redis_rate/v8"
)

// Service provides various useful HTTP and GRPC middlewares, eg. for authentication and rate limiting
type Service struct {
	Redis         *redis.Client
	SecretService *secret.Service

	limiter *redis_rate.Limiter
}

// New creates a new middleware service
func New(redis *redis.Client, secretService *secret.Service) *Service {
	s := &Service{
		Redis:         redis,
		SecretService: secretService,
	}
	s.initRateLimiter()
	return s
}
