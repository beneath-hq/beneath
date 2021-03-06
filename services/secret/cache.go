package secret

import (
	"context"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/go-redis/cache/v7"
	uuid "github.com/satori/go.uuid"
	"github.com/vmihailenco/msgpack"

	"github.com/beneath-hq/beneath/models"
	"github.com/beneath-hq/beneath/pkg/secrettoken"
)

// AuthenticateWithToken returns the secret object matching token, or nil.
// NOTE: It does *not* return AnonymousSecret when token is empty. The auth middleware
// returns AnonymousSecret if no token is provided, but this function checks a provided token.
func (s *Service) AuthenticateWithToken(ctx context.Context, token secrettoken.Token) models.Secret {
	return s.cache.Get(ctx, token)
}

// Cache wraps a Redis and LRU cache of authenticated secrets
type Cache struct {
	codec   *cache.Codec
	service *Service
}

// cache configuration
var secretCacheConfig = struct {
	cacheTime    time.Duration
	cacheLRUTime time.Duration
	cacheLRUSize int
	redisKeyFn   func(hashedToken []byte) string
}{
	cacheTime:    time.Hour,
	cacheLRUTime: 10 * time.Second,
	cacheLRUSize: 10000,
	redisKeyFn: func(hashedToken []byte) string {
		return string(append([]byte("scrt:"), hashedToken...))
	},
}

// initializes the service's cache
func (s *Service) initCache() {
	s.cache = &Cache{
		codec: &cache.Codec{
			Redis:     s.Redis,
			Marshal:   msgpack.Marshal,
			Unmarshal: msgpack.Unmarshal,
		},
		service: s,
	}
	s.cache.codec.UseLocalCache(secretCacheConfig.cacheLRUSize, secretCacheConfig.cacheLRUTime)
}

// Get returns a Secret for a token in string representation
func (c *Cache) Get(ctx context.Context, token secrettoken.Token) models.Secret {
	// NOTE: explicit nil checks necessary to avoid returning a "typed nil" (where the result != nil)
	switch token.Flags() {
	case TokenFlagsUser:
		s := c.userOnce(ctx, token)
		if s == nil {
			return nil
		}
		return s
	case TokenFlagsService:
		s := c.serviceOnce(ctx, token)
		if s == nil {
			return nil
		}
		return s
	default:
		return nil
	}
}

// Clear removes a key from the cache
func (c *Cache) Clear(ctx context.Context, hashedToken []byte) {
	err := c.codec.Delete(secretCacheConfig.redisKeyFn(hashedToken))
	if err != nil && err != cache.ErrCacheMiss {
		panic(err)
	}
}

// ClearForOrganization clears all secrets for services and billing users in the org
func (c *Cache) ClearForOrganization(ctx context.Context, organizationID uuid.UUID) {
	c.clearQuery(ctx, `
		select ss.hashed_token
		from service_secrets ss
		join services s on ss.service_id = s.service_id
		join projects p on s.project_id = p.project_id
		where p.organization_id = ?
		union
		select us.hashed_token
		from user_secrets us
		join users u on us.user_id = u.user_id
		where u.billing_organization_id = ?
	`, organizationID, organizationID)
}

// ClearForUser clears all secrets for the user
func (c *Cache) ClearForUser(ctx context.Context, userID uuid.UUID) {
	c.clearQuery(ctx, `
		select us.hashed_token
		from user_secrets us
		where us.user_id = ?
	`, userID)
}

// ClearForService clears all secrets for the service
func (c *Cache) ClearForService(ctx context.Context, serviceID uuid.UUID) {
	c.clearQuery(ctx, `
		select ss.hashed_token
		from service_secrets ss
		where ss.service_id = ?
	`, serviceID)
}

// clearQuery clears hashed tokens returned by the given query and params
func (c *Cache) clearQuery(ctx context.Context, query string, params ...interface{}) {
	var hashedTokens [][]byte
	_, err := c.service.DB.GetDB(ctx).QueryContext(ctx, &hashedTokens, query, params...)
	if err != nil {
		panic(err)
	}
	for _, hashedToken := range hashedTokens {
		c.Clear(ctx, hashedToken)
	}
}

func (c *Cache) userOnce(ctx context.Context, token secrettoken.Token) *models.UserSecret {
	hashedToken := token.Hashed()
	var res *models.UserSecret
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

func (c *Cache) serviceOnce(ctx context.Context, token secrettoken.Token) *models.ServiceSecret {
	hashedToken := token.Hashed()
	var res *models.ServiceSecret
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

func (c *Cache) userGetter(ctx context.Context, hashedToken []byte) func() (interface{}, error) {
	return func() (interface{}, error) {
		secret := &models.UserSecret{}
		err := c.service.DB.GetDB(ctx).ModelContext(ctx, secret).
			Column(
				"user_secret.user_secret_id",
				"user_secret.read_only",
				"user_secret.public_only",
				"User.user_id",
				"User.quota_epoch",
				"User.read_quota",
				"User.write_quota",
				"User.scan_quota",
				"User.master",
				"User.BillingOrganization.organization_id",
				"User.BillingOrganization.quota_epoch",
				"User.BillingOrganization.read_quota",
				"User.BillingOrganization.write_quota",
				"User.BillingOrganization.scan_quota",
			).
			Where("hashed_token = ?", hashedToken).
			Select()
		if err != nil && err != pg.ErrNoRows {
			panic(err)
		}

		if secret.User != nil {
			secret.UserID = secret.User.UserID
			secret.Master = secret.User.Master
			secret.BillingQuotaEpoch = secret.User.BillingOrganization.QuotaEpoch
			secret.BillingOrganizationID = secret.User.BillingOrganization.OrganizationID
			secret.BillingReadQuota = secret.User.BillingOrganization.ReadQuota
			secret.BillingWriteQuota = secret.User.BillingOrganization.WriteQuota
			secret.BillingScanQuota = secret.User.BillingOrganization.ScanQuota
			secret.OwnerQuotaEpoch = secret.User.QuotaEpoch
			secret.OwnerReadQuota = secret.User.ReadQuota
			secret.OwnerWriteQuota = secret.User.WriteQuota
			secret.OwnerScanQuota = secret.User.ScanQuota
			secret.User = nil
		}

		return secret, nil
	}
}

func (c *Cache) serviceGetter(ctx context.Context, hashedToken []byte) func() (interface{}, error) {
	return func() (interface{}, error) {
		secret := &models.ServiceSecret{}
		err := c.service.DB.GetDB(ctx).ModelContext(ctx, secret).
			Column(
				"service_secret.service_secret_id",
				"Service.service_id",
				"Service.quota_epoch",
				"Service.read_quota",
				"Service.write_quota",
				"Service.scan_quota",
				"Service.Project.organization_id",
				"Service.Project.Organization.quota_epoch",
				"Service.Project.Organization.read_quota",
				"Service.Project.Organization.write_quota",
				"Service.Project.Organization.scan_quota",
			).
			Where("hashed_token = ?", hashedToken).
			Select()
		if err != nil && err != pg.ErrNoRows {
			panic(err)
		}

		if secret.Service != nil {
			secret.ServiceID = secret.Service.ServiceID
			secret.Master = false
			secret.BillingOrganizationID = secret.Service.Project.OrganizationID
			secret.BillingQuotaEpoch = secret.Service.Project.Organization.QuotaEpoch
			secret.BillingReadQuota = secret.Service.Project.Organization.ReadQuota
			secret.BillingWriteQuota = secret.Service.Project.Organization.WriteQuota
			secret.BillingScanQuota = secret.Service.Project.Organization.ScanQuota
			secret.OwnerQuotaEpoch = secret.Service.QuotaEpoch
			secret.OwnerReadQuota = secret.Service.ReadQuota
			secret.OwnerWriteQuota = secret.Service.WriteQuota
			secret.OwnerScanQuota = secret.Service.ScanQuota
			secret.Service = nil
		}

		return secret, nil
	}
}
