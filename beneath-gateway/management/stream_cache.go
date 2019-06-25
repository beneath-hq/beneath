package management

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/bluele/gcache"

	"github.com/go-redis/cache"
	uuid "github.com/satori/go.uuid"
)

var streamCacheConfig = struct {
	cacheTime    time.Duration
	cacheLRUTime time.Duration
	cacheLRUSize int
	redisKeyFn   func(instanceID uuid.UUID) string
}{
	cacheTime:    time.Hour,
	cacheLRUTime: time.Hour,
	cacheLRUSize: 10000,
	redisKeyFn: func(instanceID uuid.UUID) string {
		return fmt.Sprintf("stream:%s", instanceID)
	},
}

// CachedStream keeps key information about a stream for rapid lookup
type CachedStream struct {
	Public     bool
	Manual     bool
	ProjectID  uuid.UUID
	AvroSchema string
}

type internalCachedStream struct {
	Public     bool
	Manual     bool
	ProjectID  uuid.UUID
	AvroSchema string
}

// MarshalBinary serializes for storage in cache
func (c CachedStream) MarshalBinary() ([]byte, error) {
	wrapped := internalCachedStream{
		Public:     c.Public,
		Manual:     c.Manual,
		ProjectID:  c.ProjectID,
		AvroSchema: c.AvroSchema,
	}
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(wrapped)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// UnmarshalBinary deserializes back from storage in cache
func (c *CachedStream) UnmarshalBinary(data []byte) error {
	wrapped := internalCachedStream{}
	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&wrapped)
	if err != nil {
		return err
	}

	c.Public = wrapped.Public
	c.Manual = wrapped.Manual
	c.ProjectID = wrapped.ProjectID
	c.AvroSchema = wrapped.AvroSchema

	return nil
}

// StreamCache is a Redis and LRU based cache mapping an instance ID to a CachedStream
type StreamCache struct {
	codec *cache.Codec
	lru   gcache.Cache
}

// NewStreamCache returns a new StreamCache
func NewStreamCache() (c StreamCache) {
	c.codec = &cache.Codec{
		Redis:     Redis,
		Marshal:   c.marshal,
		Unmarshal: c.unmarshal,
	}

	c.lru = gcache.New(streamCacheConfig.cacheLRUSize).LRU().Build()

	return c
}

// Get returns the current instance ID for the specified stream
func (c StreamCache) Get(instanceID uuid.UUID) (cachedStream *CachedStream, err error) {
	key := streamCacheConfig.redisKeyFn(instanceID)

	// lookup in lru first
	value, err := c.lru.Get(key)
	if err == nil {
		cachedStream = value.(*CachedStream)
		return cachedStream, nil
	}

	// lookup in redis or db
	cachedStream = &CachedStream{}
	err = c.codec.Once(&cache.Item{
		Key:        key,
		Object:     cachedStream,
		Expiration: streamCacheConfig.cacheTime,
		Func:       c.getterFunc(instanceID),
	})

	if err != nil {
		log.Panicf("StreamCache.Get error: %v", err)
	}

	if cachedStream.ProjectID == uuid.Nil {
		cachedStream = nil
		err = errors.New("stream not found")
	}

	// set in lru
	c.lru.SetWithExpire(key, cachedStream, streamCacheConfig.cacheLRUTime)

	return cachedStream, err
}

func (c StreamCache) marshal(v interface{}) ([]byte, error) {
	cachedStream := v.(*CachedStream)
	return cachedStream.MarshalBinary()
}

func (c StreamCache) unmarshal(b []byte, v interface{}) (err error) {
	cachedStream := v.(*CachedStream)
	return cachedStream.UnmarshalBinary(b)
}

func (c StreamCache) getterFunc(instanceID uuid.UUID) func() (interface{}, error) {
	return func() (interface{}, error) {
		res := &CachedStream{}
		_, err := DB.Query(res, `
				select p.public, s.manual, s.project_id, s.avro_schema
				from stream_instances si
				join streams s on si.stream_id = s.stream_id
				join projects p on s.project_id = p.project_id
				where si.stream_instance_id = ?
			`, instanceID)
		return res, err
	}
}
