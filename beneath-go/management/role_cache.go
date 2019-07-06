package management

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/go-pg/pg"
	"github.com/go-redis/cache"
	uuid "github.com/satori/go.uuid"
)

var roleCacheConfig = struct {
	cacheTime    time.Duration
	cacheLRUTime time.Duration
	cacheLRUSize int
	redisKeyFn   func(key string, projectID uuid.UUID) string
}{
	cacheTime:    time.Hour,
	cacheLRUTime: 1 * time.Minute,
	cacheLRUSize: 20000,
	redisKeyFn: func(key string, projectID uuid.UUID) string {
		return fmt.Sprintf("role:%s:%s", key, projectID)
	},
}

// CachedRole describes a key's permissions
type CachedRole struct {
	ReadPublic bool
	Read       bool
	Write      bool
	Manage     bool
}

// MarshalBinary serializes for storage in cache
func (c CachedRole) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(c)
	if err != nil {
		return nil, fmt.Errorf("marshal: encoding CachedRole failed: %v", err)
	}
	return buf.Bytes(), nil
}

// UnmarshalBinary deserializes back from storage in cache
func (c *CachedRole) UnmarshalBinary(data []byte) error {
	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(c)
	if err != nil {
		return fmt.Errorf("unmarshal: decoding CachedRole failed: %v", err)
	}
	return nil
}

// RoleCache is a Redis and LRU based cache mapping (hashedKey, projectID) pairs to CachedRole
type RoleCache struct {
	codec *cache.Codec
}

// NewRoleCache returns a new RoleCache
func NewRoleCache() (c RoleCache) {
	c.codec = &cache.Codec{
		Redis:     Redis,
		Marshal:   c.marshal,
		Unmarshal: c.unmarshal,
	}

	c.codec.UseLocalCache(roleCacheConfig.cacheLRUSize, roleCacheConfig.cacheLRUTime)

	return c
}

// Get returns the role for key in project
func (c RoleCache) Get(key string, projectID uuid.UUID) (cachedRole CachedRole, err error) {
	err = c.codec.Once(&cache.Item{
		Key:        roleCacheConfig.redisKeyFn(key, projectID),
		Object:     &cachedRole,
		Expiration: roleCacheConfig.cacheTime,
		Func:       c.getterFunc(key, projectID),
	})

	if err != nil {
		log.Panicf("RoleCache.Get error: %v", err)
	}

	if !cachedRole.ReadPublic && !cachedRole.Read && !cachedRole.Write && !cachedRole.Manage {
		err = errors.New("key not found")
	}

	return cachedRole, err
}

func (c RoleCache) marshal(v interface{}) ([]byte, error) {
	cachedRole := v.(CachedRole)
	return cachedRole.MarshalBinary()
}

func (c RoleCache) unmarshal(b []byte, v interface{}) (err error) {
	cachedRole := v.(*CachedRole)
	return cachedRole.UnmarshalBinary(b)
}

func (c RoleCache) getterFunc(key string, projectID uuid.UUID) func() (interface{}, error) {
	return func() (interface{}, error) {
		res := ""
		_, err := DB.Query(pg.Scan(&res), `
			select coalesce(
				(select k.role
					from keys k
					where k.hashed_key = ?0 and k.project_id = ?1
				),
				(select 'm' as role
					from keys k
					join projects_users pu on k.user_id is not null and k.user_id = pu.user_id
					where k.hashed_key = ?0 and pu.project_id = ?1
				),
				(select '-' as role
					from keys k
					where k.hashed_key = ?0
				),
				''
			)
		`, key, projectID)

		if err != nil {
			return nil, err
		}

		return CachedRole{
			ReadPublic: res != "",
			Read:       res == "r" || res == "rw",
			Write:      res == "rw",
			Manage:     res == "m",
		}, nil
	}
}
