package entity

import (
	"context"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/go-redis/cache/v7"
	uuid "github.com/satori/go.uuid"
	"github.com/vmihailenco/msgpack"

	"gitlab.com/beneath-hq/beneath/internal/hub"
	"gitlab.com/beneath-hq/beneath/pkg/secrettoken"
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
		redisKeyFn   func(hashedToken []byte) string
	}{
		cacheTime:    time.Hour,
		cacheLRUTime: 1 * time.Minute,
		cacheLRUSize: 10000,
		redisKeyFn: func(hashedToken []byte) string {
			return string(append([]byte("scrt:"), hashedToken...))
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
				Redis:     hub.Redis,
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

// Clear removes a key from the cache
func (c *SecretCache) Clear(ctx context.Context, hashedToken []byte) {
	err := c.codec.Delete(secretCacheConfig.redisKeyFn(hashedToken))
	if err != nil && err != cache.ErrCacheMiss {
		panic(err)
	}
}

// ClearForOrganization clears all secrets for services and billing users in the org
func (c *SecretCache) ClearForOrganization(ctx context.Context, organizationID uuid.UUID) {
	c.clearQuery(ctx, `
		select ss.hashed_token
		from service_secrets ss
		join services s on ss.service_id = s.service_id
		where s.organization_id = ?
		union
		select us.hashed_token
		from user_secrets us
		join users u on us.user_id = u.user_id
		where u.billing_organization_id = ?
	`, organizationID, organizationID)
}

// ClearForUser clears all secrets for the user
func (c *SecretCache) ClearForUser(ctx context.Context, userID uuid.UUID) {
	c.clearQuery(ctx, `
		select us.hashed_token
		from user_secrets us
		where us.user_id = ?
	`, userID)
}

// ClearForService clears all secrets for the service
func (c *SecretCache) ClearForService(ctx context.Context, serviceID uuid.UUID) {
	c.clearQuery(ctx, `
		select ss.hashed_token
		from service_secrets ss
		where ss.service_id = ?
	`, serviceID)
}

// clearQuery clears hashed tokens returned by the given query and params
func (c *SecretCache) clearQuery(ctx context.Context, query string, params ...interface{}) {
	var hashedTokens [][]byte
	_, err := hub.DB.QueryContext(ctx, &hashedTokens, query, params...)
	if err != nil {
		panic(err)
	}
	for _, hashedToken := range hashedTokens {
		c.Clear(ctx, hashedToken)
	}
}

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

func (c *SecretCache) userGetter(ctx context.Context, hashedToken []byte) func() (interface{}, error) {
	return func() (interface{}, error) {
		secret := &UserSecret{}
		err := hub.DB.ModelContext(ctx, secret).
			Column(
				"user_secret.user_secret_id",
				"user_secret.read_only",
				"user_secret.public_only",
				"User.user_id",
				"User.read_quota",
				"User.write_quota",
				"User.master",
				"User.BillingOrganization.organization_id",
				"User.BillingOrganization.read_quota",
				"User.BillingOrganization.write_quota",
			).
			Where("hashed_token = ?", hashedToken).
			Select()
		if err != nil && err != pg.ErrNoRows {
			panic(err)
		}

		if secret.User != nil {
			secret.UserID = secret.User.UserID
			secret.Master = secret.User.Master
			secret.BillingOrganizationID = secret.User.BillingOrganization.OrganizationID
			secret.BillingReadQuota = secret.User.BillingOrganization.ReadQuota
			secret.BillingWriteQuota = secret.User.BillingOrganization.WriteQuota
			secret.OwnerReadQuota = secret.User.ReadQuota
			secret.OwnerWriteQuota = secret.User.WriteQuota
			secret.User = nil
		}

		return secret, nil
	}
}

func (c *SecretCache) serviceGetter(ctx context.Context, hashedToken []byte) func() (interface{}, error) {
	return func() (interface{}, error) {
		secret := &ServiceSecret{}
		err := hub.DB.ModelContext(ctx, secret).
			Column(
				"service_secret.service_secret_id",
				"Service.service_id",
				"Service.read_quota",
				"Service.write_quota",
				"Service.Organization.organization_id",
				"Service.Organization.read_quota",
				"Service.Organization.write_quota",
			).
			Where("hashed_token = ?", hashedToken).
			Select()
		if err != nil && err != pg.ErrNoRows {
			panic(err)
		}

		if secret.Service != nil {
			secret.ServiceID = secret.Service.ServiceID
			secret.Master = false
			secret.BillingOrganizationID = secret.Service.Organization.OrganizationID
			secret.BillingReadQuota = secret.Service.Organization.ReadQuota
			secret.BillingWriteQuota = secret.Service.Organization.WriteQuota
			secret.OwnerReadQuota = secret.Service.ReadQuota
			secret.OwnerWriteQuota = secret.Service.WriteQuota
			secret.Service = nil
		}

		return secret, nil
	}
}
