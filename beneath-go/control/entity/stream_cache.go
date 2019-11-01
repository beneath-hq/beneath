package entity

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"sync"
	"time"

	"github.com/beneath-core/beneath-go/core/codec"
	"github.com/beneath-core/beneath-go/db"
	"github.com/bluele/gcache"
	"github.com/go-pg/pg/v9"
	"github.com/go-redis/cache/v7"
	uuid "github.com/satori/go.uuid"
)

// CachedStream keeps key information about a stream for rapid lookup
type CachedStream struct {
	StreamID    uuid.UUID
	Public      bool
	External    bool
	Batch       bool
	Manual      bool
	Committed   bool
	ProjectID   uuid.UUID
	ProjectName string
	StreamName  string
	Codec       *codec.Codec
}

type internalCachedStream struct {
	StreamID            uuid.UUID
	Public              bool
	External            bool
	Batch               bool
	Manual              bool
	Committed           bool
	ProjectID           uuid.UUID
	ProjectName         string
	StreamName          string
	KeyFields           []string
	CanonicalAvroSchema string
}

// MarshalBinary serializes for storage in cache
func (c CachedStream) MarshalBinary() ([]byte, error) {
	wrapped := internalCachedStream{
		StreamID:    c.StreamID,
		Public:      c.Public,
		External:    c.External,
		Batch:       c.Batch,
		Manual:      c.Manual,
		Committed:   c.Committed,
		ProjectID:   c.ProjectID,
		ProjectName: c.ProjectName,
		StreamName:  c.StreamName,
	}

	// necessary because we allow empty CachedStream objects
	if c.Codec != nil {
		wrapped.KeyFields = c.Codec.GetKeyFields()
		wrapped.CanonicalAvroSchema = c.Codec.GetAvroSchemaString()
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

	return unwrapInternalCachedStream(&wrapped, c)
}

// streamCache is a Redis and LRU based cache mapping an instance ID to a CachedStream
type streamCache struct {
	codec *cache.Codec
	lru   gcache.Cache
}

var (
	_streamCacheLock sync.Mutex
	_streamCache     streamCache
)

// getStreamCache returns a global streamCache
func getStreamCache() streamCache {
	_streamCacheLock.Lock()
	if _streamCache.codec == nil {
		_streamCache.codec = &cache.Codec{
			Redis:     db.Redis,
			Marshal:   _streamCache.marshal,
			Unmarshal: _streamCache.unmarshal,
		}
		_streamCache.lru = gcache.New(_streamCache.cacheLRUSize()).LRU().Build()
	}
	_streamCacheLock.Unlock()
	return _streamCache
}

// get returns the CachedStream for the given instanceID
func (c streamCache) get(ctx context.Context, instanceID uuid.UUID) *CachedStream {
	key := c.redisKey(instanceID)

	// lookup in lru first
	value, err := c.lru.Get(key)
	if err == nil {
		cachedStream := value.(*CachedStream)
		return cachedStream
	}

	// lookup in redis or db
	cachedStream := &CachedStream{}
	err = c.codec.Once(&cache.Item{
		Key:        key,
		Object:     cachedStream,
		Expiration: c.cacheTime(),
		Func:       c.getterFunc(ctx, instanceID),
	})

	if err != nil {
		panic(err)
	}

	if cachedStream.ProjectID == uuid.Nil {
		cachedStream = nil
	}

	// set in lru
	c.lru.SetWithExpire(key, cachedStream, c.cacheLRUTime())

	return cachedStream
}

func (c streamCache) clear(ctx context.Context, instanceID uuid.UUID) {
	key := c.redisKey(instanceID)
	c.lru.Remove(key)
	err := c.codec.Delete(key)
	if err != nil && err != cache.ErrCacheMiss {
		panic(err)
	}
}

func (c streamCache) cacheTime() time.Duration {
	return time.Hour
}

func (c streamCache) cacheLRUSize() int {
	return 10000
}

func (c streamCache) cacheLRUTime() time.Duration {
	return time.Minute
}

func (c streamCache) redisKey(instanceID uuid.UUID) string {
	return fmt.Sprintf("stream:%s", instanceID.String())
}

func (c streamCache) marshal(v interface{}) ([]byte, error) {
	cachedStream := v.(*CachedStream)
	return cachedStream.MarshalBinary()
}

func (c streamCache) unmarshal(b []byte, v interface{}) (err error) {
	cachedStream := v.(*CachedStream)
	return cachedStream.UnmarshalBinary(b)
}

func (c streamCache) getterFunc(ctx context.Context, instanceID uuid.UUID) func() (interface{}, error) {
	return func() (interface{}, error) {
		internalResult := &internalCachedStream{}
		_, err := db.DB.QueryContext(ctx, internalResult, `
				select
					s.stream_id,
					p.public,
					s.external,
					s.batch,
					s.manual,
					si.committed_on is not null as committed,
					s.project_id,
					p.name as project_name,
					s.name as stream_name,
					s.key_fields,
					s.canonical_avro_schema
				from stream_instances si
				join streams s on si.stream_id = s.stream_id
				join projects p on s.project_id = p.project_id
				where si.stream_instance_id = ?
			`, instanceID)

		result := &CachedStream{}
		if err == pg.ErrNoRows {
			return result, nil
		} else if err != nil {
			return nil, err
		}

		unwrapInternalCachedStream(internalResult, result)

		return result, nil
	}
}

func unwrapInternalCachedStream(source *internalCachedStream, target *CachedStream) (err error) {
	target.StreamID = source.StreamID
	target.Public = source.Public
	target.External = source.External
	target.Batch = source.Batch
	target.Manual = source.Manual
	target.Committed = source.Committed
	target.ProjectID = source.ProjectID
	target.ProjectName = source.ProjectName
	target.StreamName = source.StreamName

	// nil checks necessary because we allow empty CachedStream objects

	if source.CanonicalAvroSchema != "" && len(source.KeyFields) != 0 {
		target.Codec, err = codec.New(source.CanonicalAvroSchema, source.KeyFields)
		if err != nil {
			return err
		}
	}

	return nil
}
