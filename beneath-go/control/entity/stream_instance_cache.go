package entity

import (
	"context"
	"fmt"
	"time"

	"github.com/beneath-core/beneath-go/db"
	"github.com/go-pg/pg"
	"github.com/go-redis/cache"
	uuid "github.com/satori/go.uuid"
)

// instanceCache is a Redis and LRU based cache mapping (projectName, streamName) pairs to
// the current instance ID for that stream (table `streams`, column `current_stream_instance_id`)
type instanceCache struct {
	codec *cache.Codec
}

var (
	_instanceCache instanceCache
)

// getInstanceCache returns a global instanceCache
func getInstanceCache() instanceCache {
	if _instanceCache.codec == nil {
		_instanceCache.codec = &cache.Codec{
			Redis:     db.Redis,
			Marshal:   _instanceCache.marshal,
			Unmarshal: _instanceCache.unmarshal,
		}
		_instanceCache.codec.UseLocalCache(_instanceCache.cacheLRUSize(), _instanceCache.cacheLRUTime())
	}

	return _instanceCache
}

func (c instanceCache) get(ctx context.Context, streamName string, projectName string) uuid.UUID {
	var instanceID uuid.UUID
	err := c.codec.Once(&cache.Item{
		Key:        c.redisKey(streamName, projectName),
		Object:     &instanceID,
		Expiration: c.cacheTime(),
		Func:       c.getterFunc(ctx, streamName, projectName),
	})

	if err != nil {
		panic(err)
	}

	return instanceID
}

func (c instanceCache) cacheTime() time.Duration {
	return time.Hour
}

func (c instanceCache) cacheLRUSize() int {
	return 10000
}

func (c instanceCache) cacheLRUTime() time.Duration {
	return 1 * time.Minute
}

func (c instanceCache) redisKey(streamName string, projectName string) string {
	return fmt.Sprintf("instance_id:%s:%s", projectName, streamName)
}

func (c instanceCache) marshal(v interface{}) ([]byte, error) {
	instanceID := v.(uuid.UUID)
	return instanceID.Bytes(), nil
}

func (c instanceCache) unmarshal(b []byte, v interface{}) (err error) {
	instanceID := v.(*uuid.UUID)
	*instanceID, err = uuid.FromBytes(b)
	return err
}

func (c instanceCache) getterFunc(ctx context.Context, streamName string, projectName string) func() (interface{}, error) {
	return func() (interface{}, error) {
		res := uuid.Nil
		_, err := db.DB.QueryContext(ctx, pg.Scan(&res), `
			select s.current_stream_instance_id
			from streams s
			join projects p on s.project_id = p.project_id
			where lower(s.name) = lower(?) and lower(p.name) = lower(?)
		`, streamName, projectName)
		if err == pg.ErrNoRows {
			return res, nil
		}
		return res, err
	}
}
