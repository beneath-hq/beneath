package entity

import (
	"context"
	"fmt"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	"github.com/go-redis/cache/v7"
	uuid "github.com/satori/go.uuid"
	"github.com/vmihailenco/msgpack"

	"github.com/beneath-core/db"
	"github.com/beneath-core/pkg/secrettoken"
)

// SecretCache encapsulates a Redis cache of authenticated secrets
type SecretCache struct {
	codec cache.Codec
}

var (
	// cache global
	secretCache *SecretCache

	// configuration global
	secretCacheConfig = struct {
		cacheTime    time.Duration
		cacheLRUTime time.Duration
		cacheLRUSize int
		redisKeyFn   func(hashedSecret string) string
	}{
		cacheTime:    time.Hour,
		cacheLRUTime: 1 * time.Minute,
		cacheLRUSize: 10000,
		redisKeyFn: func(hashedSecret string) string {
			return fmt.Sprintf("secret:%s", hashedSecret)
		},
	}
)

// AuthenticateWithToken returns the secret object matching token or AnonymousSecret
func AuthenticateWithToken(ctx context.Context, token secrettoken.Token) Secret {
	return getSecretCache().Get(ctx, token)
}

func getSecretCache() *SecretCache {
	if secretCache == nil {
		secretCache = &SecretCache{
			codec: cache.Codec{
				Redis:     db.Redis,
				Marshal:   msgpack.Marshal,
				Unmarshal: msgpack.Unmarshal,
			},
		}
		secretCache.codec.UseLocalCache(secretCacheConfig.cacheLRUSize, secretCacheConfig.cacheLRUTime)
	}
	return secretCache
}

// Get returns a Secret for a token in string representation
func (c *SecretCache) Get(ctx context.Context, token secrettoken.Token) Secret {
	switch token.Flags() {
	case TokenFlagsUser:
		return c.userOnce(ctx, token)
	case TokenFlagsService:
		return c.serviceOnce(ctx, token)
	default:
		return nil
	}
}

//
func (c *SecretCache) userOnce(ctx context.Context, token secrettoken.Token) *UserSecret {
	hashedToken := token.Hashed()
	var res *UserSecret
	err := c.codec.Once(&cache.Item{
		Key:        secretCacheConfig.redisKeyFn(hashedToken),
		Object:     &res,
		Expiration: secretCacheConfig.cacheTime,
		Func:       c.userGetter(ctx, hashedToken),
	})
	if err != nil {
		panic(err)
	}
	// also caching empty results to avoid database overload on bad keys
	if res.UserSecretID == uuid.Nil {
		return nil
	}
	// setting because it's not cached in redis
	res.HashedToken = hashedToken
	return res
}

//
func (c *SecretCache) serviceOnce(ctx context.Context, token secrettoken.Token) *ServiceSecret {
	hashedToken := token.Hashed()
	var res *ServiceSecret
	err := c.codec.Once(&cache.Item{
		Key:        secretCacheConfig.redisKeyFn(hashedToken),
		Object:     &res,
		Expiration: secretCacheConfig.cacheTime,
		Func:       c.serviceGetter(ctx, hashedToken),
	})
	if err != nil {
		panic(err)
	}
	// also caching empty results to avoid database overload on bad keys
	if res.ServiceSecretID == uuid.Nil {
		return nil
	}
	// setting because it's not cached in redis
	res.HashedToken = hashedToken
	return res
}

// Delete removes a key from the cache
func (c *SecretCache) Delete(ctx context.Context, hashedToken string) {
	err := c.codec.Delete(secretCacheConfig.redisKeyFn(hashedToken))
	if err != nil && err != cache.ErrCacheMiss {
		panic(err)
	}
}

func (c *SecretCache) userGetter(ctx context.Context, hashedToken string) func() (interface{}, error) {
	return func() (interface{}, error) {
		secret := &UserSecret{}
		err := db.DB.ModelContext(ctx, secret).
			Relation("User", func(q *orm.Query) (*orm.Query, error) {
				return q.Column("user.read_quota", "user.write_quota"), nil
			}).
			Column("user_secret_id", "read_only", "public_only", "user_secret.user_id").
			Where("hashed_token = ?", hashedToken).
			Select()
		if err != nil && err != pg.ErrNoRows {
			panic(err)
		}

		if secret.User != nil {
			secret.ReadQuota = secret.User.ReadQuota
			secret.WriteQuota = secret.User.WriteQuota
			secret.User = nil
		}

		return secret, nil
	}
}

func (c *SecretCache) serviceGetter(ctx context.Context, hashedToken string) func() (interface{}, error) {
	return func() (interface{}, error) {
		secret := &ServiceSecret{}
		err := db.DB.ModelContext(ctx, secret).
			Relation("Service", func(q *orm.Query) (*orm.Query, error) {
				return q.Column("service.read_quota", "service.write_quota"), nil
			}).
			Column("service_secret_id", "service_secret.service_id").
			Where("hashed_token = ?", hashedToken).
			Select()
		if err != nil && err != pg.ErrNoRows {
			panic(err)
		}

		if secret.Service != nil {
			secret.ReadQuota = secret.Service.ReadQuota
			secret.WriteQuota = secret.Service.WriteQuota
			secret.Service = nil
		}

		return secret, nil
	}
}
