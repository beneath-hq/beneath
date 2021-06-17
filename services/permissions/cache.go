package permissions

import (
	"bytes"
	"context"
	"encoding/gob"
	"reflect"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/go-redis/cache/v7"
	uuid "github.com/satori/go.uuid"

	"github.com/beneath-hq/beneath/models"
)

// initCaches sets up the three permissions caches
func (s *Service) initCaches() {
	s.userOrganizationCache = NewCache(s, models.OrganizationPermissions{}, `
			select p.view, p.create, p.admin
			from permissions_users_organizations p
			where p.user_id = ? and p.organization_id = ?
		`)

	s.userProjectCache = NewCache(s, models.ProjectPermissions{}, `
			select p.view, p.create, p.admin
			from permissions_users_projects p
			where p.user_id = ? and p.project_id = ?
		`)

	s.serviceTableCache = NewCache(s, models.TablePermissions{}, `
			select p.read, p.write
			from permissions_services_tables p
			where p.service_id = ? and p.table_id = ?
		`)
}

// CachedUserOrganizationPermissions returns organization permissions for a given owner-resource combo
func (s *Service) CachedUserOrganizationPermissions(ctx context.Context, userID uuid.UUID, organizationID uuid.UUID) models.OrganizationPermissions {
	return s.userOrganizationCache.Get(ctx, userID, organizationID).(models.OrganizationPermissions)
}

// CachedUserProjectPermissions returns project permissions for a given owner-resource combo
func (s *Service) CachedUserProjectPermissions(ctx context.Context, userID uuid.UUID, projectID uuid.UUID) models.ProjectPermissions {
	return s.userProjectCache.Get(ctx, userID, projectID).(models.ProjectPermissions)
}

// CachedServiceTablePermissions returns table permissions for a given owner-resource combo
func (s *Service) CachedServiceTablePermissions(ctx context.Context, serviceID uuid.UUID, tableID uuid.UUID) models.TablePermissions {
	return s.serviceTableCache.Get(ctx, serviceID, tableID).(models.TablePermissions)
}

// Cache caches an owner's permissions for a resource for fast access
type Cache struct {
	service   *Service
	codec     *cache.Codec
	prototype reflect.Type
	query     string
}

var permsCacheConfig = struct {
	cacheTime    time.Duration
	cacheLRUTime time.Duration
	cacheLRUSize int
	redisKeyFn   func(ownerID uuid.UUID, resourceID uuid.UUID) string
}{
	cacheTime:    time.Hour,
	cacheLRUTime: 10 * time.Second,
	cacheLRUSize: 20000,
	redisKeyFn: func(ownerID uuid.UUID, resourceID uuid.UUID) string {
		res := append([]byte("perm:"), ownerID.Bytes()...)
		res = append(res, resourceID.Bytes()...)
		return string(res)
	},
}

// NewCache initializes a PermissionCache object for a given prototype (organization/project/table)
func NewCache(service *Service, prototype interface{}, query string) *Cache {
	pm := &Cache{}
	pm.service = service
	pm.prototype = reflect.TypeOf(prototype)
	pm.query = query
	pm.codec = &cache.Codec{
		Redis:     service.Redis,
		Marshal:   pm.marshal,
		Unmarshal: pm.unmarshal,
	}
	pm.codec.UseLocalCache(permsCacheConfig.cacheLRUSize, permsCacheConfig.cacheLRUTime)
	return pm
}

// Get fetches permissions by applying the cached query to the given parameters
func (c *Cache) Get(ctx context.Context, ownerID uuid.UUID, resourceID uuid.UUID) interface{} {
	res := reflect.New(c.prototype)
	err := c.codec.Once(&cache.Item{
		Key:        permsCacheConfig.redisKeyFn(ownerID, resourceID),
		Object:     res.Interface(),
		Expiration: permsCacheConfig.cacheTime,
		Func:       c.getterFunc(ctx, ownerID, resourceID),
	})
	if err != nil {
		if ctx.Err() == context.Canceled {
			return res.Elem().Interface()
		}
		panic(err)
	}
	return res.Elem().Interface()
}

// Clear removes a key from the cache
func (c *Cache) Clear(ctx context.Context, ownerID uuid.UUID, resourceID uuid.UUID) {
	err := c.codec.Delete(permsCacheConfig.redisKeyFn(ownerID, resourceID))
	if err != nil && err != cache.ErrCacheMiss {
		panic(err)
	}
}

func (c *Cache) marshal(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(v)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (c *Cache) unmarshal(b []byte, v interface{}) (err error) {
	dec := gob.NewDecoder(bytes.NewReader(b))
	err = dec.Decode(v)
	if err != nil {
		return err
	}
	return nil
}

func (c *Cache) getterFunc(ctx context.Context, ownerID uuid.UUID, resourceID uuid.UUID) func() (interface{}, error) {
	return func() (interface{}, error) {
		res := reflect.New(c.prototype)
		_, err := c.service.DB.GetDB(ctx).QueryContext(ctx, res.Interface(), c.query, ownerID, resourceID)
		if err != nil && err != pg.ErrNoRows {
			if ctx.Err() == context.Canceled {
				return res.Elem().Interface(), nil
			}
			panic(err)
		}
		return res.Elem().Interface(), nil
	}
}
