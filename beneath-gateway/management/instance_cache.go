package management

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/go-pg/pg"
	"github.com/go-redis/cache"
	uuid "github.com/satori/go.uuid"
)

var instanceCacheConfig = struct {
	cacheTime    time.Duration
	cacheLRUTime time.Duration
	cacheLRUSize int
	redisKeyFn   func(projectName string, streamName string) string
}{
	cacheTime:    time.Hour,
	cacheLRUTime: 1 * time.Minute,
	cacheLRUSize: 10000,
	redisKeyFn: func(projectName string, streamName string) string {
		return fmt.Sprintf("instance_id:%s:%s", projectName, streamName)
	},
}

// InstanceCache is a Redis and LRU based cache mapping (projectName, streamName) pairs to
// the current instance ID for that stream (table `streams`, column `current_stream_instance_id`)
type InstanceCache struct {
	codec *cache.Codec
}

// NewInstanceCache returns a new InstanceCache
func NewInstanceCache() (c InstanceCache) {
	c.codec = &cache.Codec{
		Redis:     Redis,
		Marshal:   c.marshal,
		Unmarshal: c.unmarshal,
	}

	c.codec.UseLocalCache(instanceCacheConfig.cacheLRUSize, instanceCacheConfig.cacheLRUTime)

	return c
}

// Get returns the current instance ID for the specified stream
func (c InstanceCache) Get(projectName string, streamName string) (instanceID uuid.UUID, err error) {
	err = c.codec.Once(&cache.Item{
		Key:        instanceCacheConfig.redisKeyFn(projectName, streamName),
		Object:     &instanceID,
		Expiration: instanceCacheConfig.cacheTime,
		Func:       c.getterFunc(projectName, streamName),
	})

	if err != nil {
		log.Panicf("InstanceCache.Get error: %v", err)
	}

	if instanceID == uuid.Nil {
		err = errors.New("stream not found")
	}

	return instanceID, err
}

func (c InstanceCache) marshal(v interface{}) ([]byte, error) {
	instanceID := v.(uuid.UUID)
	return instanceID.Bytes(), nil
}

func (c InstanceCache) unmarshal(b []byte, v interface{}) (err error) {
	instanceID := v.(*uuid.UUID)
	*instanceID, err = uuid.FromBytes(b)
	return err
}

func (c InstanceCache) getterFunc(projectName string, streamName string) func() (interface{}, error) {
	return func() (interface{}, error) {
		res := uuid.Nil
		_, err := DB.Query(pg.Scan(&res), `
			select s.current_stream_instance_id
			from streams s
			join projects p on s.project_id = p.project_id
			where s.name = lower(?) and p.name = lower(?)
		`, streamName, projectName)
		return res, err
	}
}
