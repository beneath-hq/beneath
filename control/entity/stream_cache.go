package entity

import (
	"bytes"
	"context"
	"encoding/gob"
	"sync"
	"time"

	"github.com/bluele/gcache"
	"github.com/go-pg/pg/v9"
	"github.com/go-redis/cache/v7"
	uuid "github.com/satori/go.uuid"
	"github.com/vmihailenco/msgpack"

	"gitlab.com/beneath-hq/beneath/internal/hub"
	"gitlab.com/beneath-hq/beneath/pkg/codec"
)

// FindCachedStreamByCurrentInstanceID returns select info about the instance's stream
func FindCachedStreamByCurrentInstanceID(ctx context.Context, instanceID uuid.UUID) *CachedStream {
	return getStreamCache().Get(ctx, instanceID)
}

// CachedStream keeps key information about a stream for rapid lookup
type CachedStream struct {
	StreamID                  uuid.UUID
	Public                    bool
	Final                     bool
	UseLog                    bool
	UseIndex                  bool
	UseWarehouse              bool
	LogRetentionSeconds       int32
	IndexRetentionSeconds     int32
	WarehouseRetentionSeconds int32
	OrganizationID            uuid.UUID
	OrganizationName          string
	ProjectID                 uuid.UUID
	ProjectName               string
	StreamName                string
	Codec                     *codec.Codec
}

// EfficientStreamIndex represents indexes in EfficientStream
type EfficientStreamIndex struct {
	StreamIndexID uuid.UUID
	ShortID       int
	Fields        []string
	Primary       bool
	Normalize     bool
}

// GetIndexID implements codec.Index
func (e EfficientStreamIndex) GetIndexID() uuid.UUID {
	return e.StreamIndexID
}

// GetShortID implements codec.Index
func (e EfficientStreamIndex) GetShortID() int {
	return e.ShortID
}

// GetFields implements codec.Index
func (e EfficientStreamIndex) GetFields() []string {
	return e.Fields
}

// GetNormalize implements codec.Index
func (e EfficientStreamIndex) GetNormalize() bool {
	return e.Normalize
}

type internalCachedStream struct {
	StreamID                  uuid.UUID
	Public                    bool
	Final                     bool
	UseLog                    bool
	UseIndex                  bool
	UseWarehouse              bool
	LogRetentionSeconds       int32
	IndexRetentionSeconds     int32
	WarehouseRetentionSeconds int32
	OrganizationID            uuid.UUID
	OrganizationName          string
	ProjectID                 uuid.UUID
	ProjectName               string
	StreamName                string
	CanonicalAvroSchema       string
	Indexes                   []EfficientStreamIndex
}

// NewCachedStream creates a CachedStream from a regular Stream object
func NewCachedStream(s *Stream, instance *StreamInstance) *CachedStream {
	if s.Project == nil {
		panic("Stream project must be loaded")
	}

	if s.Project.Organization == nil {
		panic("Stream organization must be loaded")
	}

	indexes := make([]EfficientStreamIndex, len(s.StreamIndexes))
	for idx, index := range s.StreamIndexes {
		indexes[idx] = EfficientStreamIndex{
			StreamIndexID: index.StreamIndexID,
			ShortID:       index.ShortID,
			Fields:        index.Fields,
			Primary:       index.Primary,
			Normalize:     index.Normalize,
		}
	}

	internal := &internalCachedStream{
		StreamID:                  s.StreamID,
		Public:                    s.Project.Public,
		Final:                     instance.MadeFinalOn != nil,
		UseLog:                    s.UseLog,
		UseIndex:                  s.UseIndex,
		UseWarehouse:              s.UseWarehouse,
		LogRetentionSeconds:       s.LogRetentionSeconds,
		IndexRetentionSeconds:     s.IndexRetentionSeconds,
		WarehouseRetentionSeconds: s.WarehouseRetentionSeconds,
		OrganizationID:            s.Project.Organization.OrganizationID,
		OrganizationName:          s.Project.Organization.Name,
		ProjectID:                 s.Project.ProjectID,
		ProjectName:               s.Project.Name,
		StreamName:                s.Name,
		CanonicalAvroSchema:       s.CanonicalAvroSchema,
		Indexes:                   indexes,
	}

	result := &CachedStream{}

	err := unwrapInternalCachedStream(internal, result)
	if err != nil {
		panic(err)
	}

	return result
}

// MarshalBinary serializes for storage in cache
func (c CachedStream) MarshalBinary() ([]byte, error) {
	wrapped := internalCachedStream{
		StreamID:                  c.StreamID,
		Public:                    c.Public,
		Final:                     c.Final,
		UseLog:                    c.UseLog,
		UseIndex:                  c.UseIndex,
		UseWarehouse:              c.UseWarehouse,
		LogRetentionSeconds:       c.LogRetentionSeconds,
		IndexRetentionSeconds:     c.IndexRetentionSeconds,
		WarehouseRetentionSeconds: c.WarehouseRetentionSeconds,
		OrganizationID:            c.OrganizationID,
		OrganizationName:          c.OrganizationName,
		ProjectID:                 c.ProjectID,
		ProjectName:               c.ProjectName,
		StreamName:                c.StreamName,
	}

	// necessary because we allow empty CachedStream objects
	if c.Codec != nil {
		wrapped.CanonicalAvroSchema = c.Codec.AvroSchema
		wrapped.Indexes = []EfficientStreamIndex{c.Codec.PrimaryIndex.(EfficientStreamIndex)}
		for _, index := range c.Codec.SecondaryIndexes {
			wrapped.Indexes = append(wrapped.Indexes, index.(EfficientStreamIndex))
		}
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

// Add support for msgpack
var _ msgpack.CustomEncoder = (*CachedStream)(nil)
var _ msgpack.CustomDecoder = (*CachedStream)(nil)

// EncodeMsgpack adds support for encoding CachedStream with msgpack
func (c *CachedStream) EncodeMsgpack(enc *msgpack.Encoder) error {
	bin, err := c.MarshalBinary()
	if err != nil {
		return err
	}
	return enc.EncodeBytes(bin)
}

// DecodeMsgpack adds support for decoding CachedStream with msgpack
func (c *CachedStream) DecodeMsgpack(dec *msgpack.Decoder) error {
	bin, err := dec.DecodeBytes()
	if err != nil {
		return err
	}
	return c.UnmarshalBinary(bin)
}

// CachedStream implements engine/driver.Organization, engine/driver.Project, engine/driver.StreamInstance, and engine/driver.Stream

// GetOrganizationID implements engine/driver.Organization
func (c *CachedStream) GetOrganizationID() uuid.UUID {
	return c.OrganizationID
}

// GetOrganizationName implements engine/driver.Project
func (c *CachedStream) GetOrganizationName() string {
	return c.OrganizationName
}

// GetProjectID implements engine/driver.Project
func (c *CachedStream) GetProjectID() uuid.UUID {
	return c.ProjectID
}

// GetProjectName implements engine/driver.Project
func (c *CachedStream) GetProjectName() string {
	return c.ProjectName
}

// GetPublic implements engine/driver.Project
func (c *CachedStream) GetPublic() bool {
	return c.Public
}

// GetStreamID implements engine/driver.Stream
func (c *CachedStream) GetStreamID() uuid.UUID {
	return c.StreamID
}

// GetStreamName implements engine/driver.Stream
func (c *CachedStream) GetStreamName() string {
	return c.StreamName
}

// GetUseLog implements engine/driver.Stream
func (c *CachedStream) GetUseLog() bool {
	return c.UseLog
}

// GetUseIndex implements engine/driver.Stream
func (c *CachedStream) GetUseIndex() bool {
	return c.UseIndex
}

// GetUseWarehouse implements engine/driver.Stream
func (c *CachedStream) GetUseWarehouse() bool {
	return c.UseWarehouse
}

// GetLogRetention implements engine/driver.Stream
func (c *CachedStream) GetLogRetention() time.Duration {
	return time.Duration(c.LogRetentionSeconds) * time.Second
}

// GetIndexRetention implements engine/driver.Stream
func (c *CachedStream) GetIndexRetention() time.Duration {
	return time.Duration(c.IndexRetentionSeconds) * time.Second
}

// GetWarehouseRetention implements engine/driver.Stream
func (c *CachedStream) GetWarehouseRetention() time.Duration {
	return time.Duration(c.WarehouseRetentionSeconds) * time.Second
}

// GetCodec implements engine/driver.Stream
func (c *CachedStream) GetCodec() *codec.Codec {
	return c.Codec
}

// StreamCache is a Redis and LRU based cache mapping an instance ID to a CachedStream
type StreamCache struct {
	codec *cache.Codec
	lru   gcache.Cache
}

var (
	_streamCacheLock sync.Mutex
	_streamCache     *StreamCache
)

// getStreamCache returns a global streamCache
func getStreamCache() *StreamCache {
	_streamCacheLock.Lock()
	if _streamCache == nil {
		_streamCache = &StreamCache{
			codec: &cache.Codec{
				Redis:     hub.Redis,
				Marshal:   _streamCache.marshal,
				Unmarshal: _streamCache.unmarshal,
			},
			lru: gcache.New(_streamCache.cacheLRUSize()).LRU().Build(),
		}
	}
	_streamCacheLock.Unlock()
	return _streamCache
}

// Get returns the CachedStream for the given instanceID
func (c *StreamCache) Get(ctx context.Context, instanceID uuid.UUID) *CachedStream {
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

// Clear removes any CachedStream cached for the given instanceID
func (c *StreamCache) Clear(ctx context.Context, instanceID uuid.UUID) {
	key := c.redisKey(instanceID)
	c.lru.Remove(key)
	err := c.codec.Delete(key)
	if err != nil && err != cache.ErrCacheMiss {
		panic(err)
	}
}

// ClearForOrganization clears all streams in the organization
func (c *StreamCache) ClearForOrganization(ctx context.Context, organizationID uuid.UUID) {
	c.clearQuery(ctx, `
		select si.stream_instance_id
		from stream_instances si
		join streams s on si.stream_id = s.stream_id
		join projects p on s.project_id = p.project_id
		where p.organization_id = ?
	`, organizationID)
}

// ClearForProject clears all streams in the project
func (c *StreamCache) ClearForProject(ctx context.Context, projectID uuid.UUID) {
	c.clearQuery(ctx, `
		select si.stream_instance_id
		from stream_instances si
		join streams s on si.stream_id = s.stream_id
		where s.project_id = ?
	`, projectID)
}

// clearQuery clears instance IDs returned by the given query and params
func (c *StreamCache) clearQuery(ctx context.Context, query string, params ...interface{}) {
	var instanceIDs []uuid.UUID
	_, err := hub.DB.QueryContext(ctx, &instanceIDs, query, params...)
	if err != nil {
		panic(err)
	}
	for _, instanceID := range instanceIDs {
		c.Clear(ctx, instanceID)
	}
}

func (c *StreamCache) cacheTime() time.Duration {
	return time.Hour
}

func (c *StreamCache) cacheLRUSize() int {
	return 10000
}

func (c *StreamCache) cacheLRUTime() time.Duration {
	return 10 * time.Second
}

func (c *StreamCache) redisKey(instanceID uuid.UUID) string {
	return string(append([]byte("strm:"), instanceID.Bytes()...))
}

func (c *StreamCache) marshal(v interface{}) ([]byte, error) {
	cachedStream := v.(*CachedStream)
	return cachedStream.MarshalBinary()
}

func (c *StreamCache) unmarshal(b []byte, v interface{}) (err error) {
	cachedStream := v.(*CachedStream)
	return cachedStream.UnmarshalBinary(b)
}

func (c *StreamCache) getterFunc(ctx context.Context, instanceID uuid.UUID) func() (interface{}, error) {
	return func() (interface{}, error) {
		internalResult := &internalCachedStream{}
		_, err := hub.DB.QueryContext(ctx, internalResult, `
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

		result := &CachedStream{}
		if err == pg.ErrNoRows {
			return result, nil
		} else if err != nil {
			return nil, err
		}

		_, err = hub.DB.QueryContext(ctx, &internalResult.Indexes, `
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
			return result, nil
		} else if err != nil {
			return nil, err
		}

		err = unwrapInternalCachedStream(internalResult, result)
		if err != nil {
			return nil, err
		}

		return result, nil
	}
}

func unwrapInternalCachedStream(source *internalCachedStream, target *CachedStream) (err error) {
	target.StreamID = source.StreamID
	target.Public = source.Public
	target.Final = source.Final
	target.UseLog = source.UseLog
	target.UseIndex = source.UseIndex
	target.UseWarehouse = source.UseWarehouse
	target.LogRetentionSeconds = source.LogRetentionSeconds
	target.IndexRetentionSeconds = source.IndexRetentionSeconds
	target.WarehouseRetentionSeconds = source.WarehouseRetentionSeconds
	target.OrganizationID = source.OrganizationID
	target.OrganizationName = source.OrganizationName
	target.ProjectID = source.ProjectID
	target.ProjectName = source.ProjectName
	target.StreamName = source.StreamName

	// nil checks necessary because we allow empty CachedStream objects

	if source.CanonicalAvroSchema != "" {
		var primaryIndex codec.Index
		var secondaryIndexes []codec.Index
		for _, index := range source.Indexes {
			if index.Primary {
				primaryIndex = index
			} else {
				secondaryIndexes = append(secondaryIndexes, index)
			}
		}

		target.Codec, err = codec.New(source.CanonicalAvroSchema, primaryIndex, secondaryIndexes)
		if err != nil {
			return err
		}
	}

	return nil
}
