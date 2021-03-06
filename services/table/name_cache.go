package table

import (
	"context"
	"fmt"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/go-redis/cache/v7"
	uuid "github.com/satori/go.uuid"
)

// FindPrimaryInstanceIDByOrganizationProjectAndName returns the current primary instance ID of the table
func (s *Service) FindPrimaryInstanceIDByOrganizationProjectAndName(ctx context.Context, organizationName string, projectName string, tableName string) uuid.UUID {
	return s.nameCache.Get(ctx, organizationName, projectName, tableName)
}

// nameCache is a Redis and LRU based cache mapping table paths to the primary instance ID for that table
type nameCache struct {
	codec   *cache.Codec
	service *Service
}

func (s *Service) initNameCache() {
	c := &nameCache{
		service: s,
	}
	c.codec = &cache.Codec{
		Redis:     s.Redis,
		Marshal:   c.marshal,
		Unmarshal: c.unmarshal,
	}
	c.codec.UseLocalCache(c.cacheLRUSize(), c.cacheLRUTime())
	s.nameCache = c
}

func (c *nameCache) Get(ctx context.Context, organizationName string, projectName string, tableName string) uuid.UUID {
	var instanceID uuid.UUID
	err := c.codec.Once(&cache.Item{
		Key:        c.redisKey(organizationName, projectName, tableName),
		Object:     &instanceID,
		Expiration: c.cacheTime(),
		Func:       c.getterFunc(ctx, organizationName, projectName, tableName),
	})

	if err != nil {
		if ctx.Err() == context.Canceled {
			return uuid.Nil
		}
		panic(err)
	}

	return instanceID
}

func (c *nameCache) Clear(ctx context.Context, organizationName string, projectName string, tableName string) {
	err := c.codec.Delete(c.redisKey(organizationName, projectName, tableName))
	if err != nil && err != cache.ErrCacheMiss {
		panic(err)
	}
}

func (c *nameCache) cacheTime() time.Duration {
	return time.Hour
}

func (c *nameCache) cacheLRUSize() int {
	return 10000
}

func (c *nameCache) cacheLRUTime() time.Duration {
	return 10 * time.Second
}

func (c *nameCache) redisKey(organizationName string, projectName string, tableName string) string {
	return fmt.Sprintf("inst:%s:%s:%s", organizationName, projectName, tableName)
}

func (c *nameCache) marshal(v interface{}) ([]byte, error) {
	instanceID := v.(uuid.UUID)
	return instanceID.Bytes(), nil
}

func (c *nameCache) unmarshal(b []byte, v interface{}) (err error) {
	instanceID := v.(*uuid.UUID)
	*instanceID, err = uuid.FromBytes(b)
	return err
}

func (c *nameCache) getterFunc(ctx context.Context, organizationName string, projectName string, tableName string) func() (interface{}, error) {
	return func() (interface{}, error) {
		res := uuid.Nil
		_, err := c.service.DB.GetDB(ctx).QueryContext(ctx, pg.Scan(&res), `
			select s.primary_table_instance_id
			from tables s
			join projects p on s.project_id = p.project_id
			join organizations o on p.organization_id = o.organization_id
			where lower(s.name) = lower(?)
			and lower(p.name) = lower(?)
			and lower(o.name) = lower(?)
			and primary_table_instance_id is not null
		`, tableName, projectName, organizationName)
		if err == pg.ErrNoRows {
			return res, nil
		}
		return res, err
	}
}
