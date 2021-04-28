package stream

import (
	"context"
	"time"

	"github.com/bluele/gcache"
	"github.com/go-pg/pg/v9"
	"github.com/go-redis/cache/v7"
	uuid "github.com/satori/go.uuid"

	"github.com/beneath-hq/beneath/models"
)

// FindCachedInstance returns select info about the instance and its stream (cached)
func (s *Service) FindCachedInstance(ctx context.Context, instanceID uuid.UUID) *models.CachedInstance {
	return s.instanceCache.Get(ctx, instanceID)
}

// StreamCache is a Redis and LRU based cache mapping an instance ID to a CachedStream
type instanceCache struct {
	codec   *cache.Codec
	lru     gcache.Cache
	service *Service
}

func (s *Service) initInstanceCache() {
	c := &instanceCache{
		service: s,
	}
	c.codec = &cache.Codec{
		Redis:     s.Redis,
		Marshal:   c.marshal,
		Unmarshal: c.unmarshal,
	}
	c.lru = gcache.New(c.cacheLRUSize()).LRU().Build()
	s.instanceCache = c
}

// Get returns the CachedStream for the given instanceID
func (c *instanceCache) Get(ctx context.Context, instanceID uuid.UUID) *models.CachedInstance {
	key := c.redisKey(instanceID)

	// lookup in lru first
	value, err := c.lru.Get(key)
	if err == nil {
		cachedStream := value.(*models.CachedInstance)
		return cachedStream
	}

	// lookup in redis or db
	cachedStream := &models.CachedInstance{}
	err = c.codec.Once(&cache.Item{
		Key:        key,
		Object:     cachedStream,
		Expiration: c.cacheTime(),
		Func:       c.getterFunc(ctx, instanceID),
	})

	if err != nil {
		if ctx.Err() == context.Canceled {
			return nil
		}
		panic(err)
	}

	if cachedStream.StreamID == uuid.Nil {
		cachedStream = nil
	}

	// set in lru
	c.lru.SetWithExpire(key, cachedStream, c.cacheLRUTime())

	return cachedStream
}

// Clear removes any CachedStream cached for the given instanceID
func (c *instanceCache) Clear(ctx context.Context, instanceID uuid.UUID) {
	key := c.redisKey(instanceID)
	c.lru.Remove(key)
	err := c.codec.Delete(key)
	if err != nil && err != cache.ErrCacheMiss {
		panic(err)
	}
}

// ClearForOrganization clears all streams in the organization
func (c *instanceCache) ClearForOrganization(ctx context.Context, organizationID uuid.UUID) {
	c.clearQuery(ctx, `
		select si.stream_instance_id
		from stream_instances si
		join streams s on si.stream_id = s.stream_id
		join projects p on s.project_id = p.project_id
		where p.organization_id = ?
	`, organizationID)
}

// ClearForProject clears all streams in the project
func (c *instanceCache) ClearForProject(ctx context.Context, projectID uuid.UUID) {
	c.clearQuery(ctx, `
		select si.stream_instance_id
		from stream_instances si
		join streams s on si.stream_id = s.stream_id
		where s.project_id = ?
	`, projectID)
}

// clearQuery clears instance IDs returned by the given query and params
func (c *instanceCache) clearQuery(ctx context.Context, query string, params ...interface{}) {
	var instanceIDs []uuid.UUID
	_, err := c.service.DB.GetDB(ctx).QueryContext(ctx, &instanceIDs, query, params...)
	if err != nil {
		panic(err)
	}
	for _, instanceID := range instanceIDs {
		c.Clear(ctx, instanceID)
	}
}

func (c *instanceCache) cacheTime() time.Duration {
	return time.Hour
}

func (c *instanceCache) cacheLRUSize() int {
	return 10000
}

func (c *instanceCache) cacheLRUTime() time.Duration {
	return 10 * time.Second
}

func (c *instanceCache) redisKey(instanceID uuid.UUID) string {
	return string(append([]byte("strm:"), instanceID.Bytes()...))
}

func (c *instanceCache) marshal(v interface{}) ([]byte, error) {
	cachedInstance := v.(*models.CachedInstance)
	return marshalCachedInstance(cachedInstance)
}

func (c *instanceCache) unmarshal(b []byte, v interface{}) (err error) {
	cachedInstance := v.(*models.CachedInstance)
	return unmarshalCachedInstance(b, cachedInstance)
}

func (c *instanceCache) getterFunc(ctx context.Context, instanceID uuid.UUID) func() (interface{}, error) {
	return func() (interface{}, error) {
		internalResult := &internalCachedInstance{}
		_, err := c.service.DB.GetDB(ctx).QueryContext(ctx, internalResult, `
				select
					s.stream_id,
					p.public,
					si.made_final_on is not null as final,
					s.use_log,
					s.use_index,
					s.use_warehouse,
					s.log_retention_seconds,
					s.index_retention_seconds,
					s.warehouse_retention_seconds,
					p.organization_id,
					o.name as organization_name,
					s.project_id,
					p.name as project_name,
					s.name as stream_name,
					s.canonical_avro_schema
				from stream_instances si
				join streams s on si.stream_id = s.stream_id
				join projects p on s.project_id = p.project_id
				join organizations o on p.organization_id = o.organization_id
				where si.stream_instance_id = ?
			`, instanceID)

		if err == pg.ErrNoRows {
			return &models.CachedInstance{}, nil
		} else if err != nil {
			return nil, err
		}

		_, err = c.service.DB.GetDB(ctx).QueryContext(ctx, &internalResult.Indexes, `
			select
				stream_index_id,
				short_id,
				fields,
				"primary",
				normalize
			from stream_indexes
			where stream_id = ?
		`, internalResult.StreamID)
		if err == pg.ErrNoRows {
			return &models.CachedInstance{}, nil
		} else if err != nil {
			return nil, err
		}

		result, err := internalToCachedInstance(internalResult)
		if err != nil {
			return nil, err
		}

		return result, nil
	}
}
