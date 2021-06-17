package table

import (
	"bytes"
	"encoding/gob"

	uuid "github.com/satori/go.uuid"

	"github.com/beneath-hq/beneath/models"
	"github.com/beneath-hq/beneath/pkg/codec"
)

// internalCachedInstance is a serializable version of models.CachedInstance (which is not serializable because of Codec)
type internalCachedInstance struct {
	TableID                   uuid.UUID
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
	TableName                 string
	CanonicalAvroSchema       string
	Indexes                   []internalCachedInstanceIndex
}

// internalCachedInstanceIndex represents indexes
type internalCachedInstanceIndex struct {
	TableIndexID uuid.UUID
	ShortID      int
	Fields       []string
	Primary      bool
	Normalize    bool
}

// GetIndexID implements codec.Index
func (i internalCachedInstanceIndex) GetIndexID() uuid.UUID {
	return i.TableIndexID
}

// GetShortID implements codec.Index
func (i internalCachedInstanceIndex) GetShortID() int {
	return i.ShortID
}

// GetFields implements codec.Index
func (i internalCachedInstanceIndex) GetFields() []string {
	return i.Fields
}

// GetNormalize implements codec.Index
func (i internalCachedInstanceIndex) GetNormalize() bool {
	return i.Normalize
}

func cachedInstanceToInternal(c *models.CachedInstance) *internalCachedInstance {
	wrapped := &internalCachedInstance{
		TableID:                   c.TableID,
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
		TableName:                 c.TableName,
	}

	// necessary because we allow empty CachedInstance objects
	if c.Codec != nil {
		wrapped.CanonicalAvroSchema = c.Codec.AvroSchema
		wrapped.Indexes = []internalCachedInstanceIndex{c.Codec.PrimaryIndex.(internalCachedInstanceIndex)}
		for _, index := range c.Codec.SecondaryIndexes {
			wrapped.Indexes = append(wrapped.Indexes, index.(internalCachedInstanceIndex))
		}
	}

	return wrapped
}

func internalToCachedInstance(i *internalCachedInstance) (*models.CachedInstance, error) {
	c := &models.CachedInstance{}

	c.TableID = i.TableID
	c.Public = i.Public
	c.Final = i.Final
	c.UseLog = i.UseLog
	c.UseIndex = i.UseIndex
	c.UseWarehouse = i.UseWarehouse
	c.LogRetentionSeconds = i.LogRetentionSeconds
	c.IndexRetentionSeconds = i.IndexRetentionSeconds
	c.WarehouseRetentionSeconds = i.WarehouseRetentionSeconds
	c.OrganizationID = i.OrganizationID
	c.OrganizationName = i.OrganizationName
	c.ProjectID = i.ProjectID
	c.ProjectName = i.ProjectName
	c.TableName = i.TableName

	// nil checks necessary because we allow empty CachedInstance objects

	if i.CanonicalAvroSchema != "" {
		var primaryIndex codec.Index
		var secondaryIndexes []codec.Index
		for _, index := range i.Indexes {
			if index.Primary {
				primaryIndex = index
			} else {
				secondaryIndexes = append(secondaryIndexes, index)
			}
		}

		codec, err := codec.New(i.CanonicalAvroSchema, primaryIndex, secondaryIndexes)
		if err != nil {
			return nil, err
		}
		c.Codec = codec
	}

	return c, nil
}

func marshalCachedInstance(c *models.CachedInstance) ([]byte, error) {
	internal := cachedInstanceToInternal(c)
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(internal)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func unmarshalCachedInstance(data []byte, target *models.CachedInstance) error {
	internal := internalCachedInstance{}
	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&internal)
	if err != nil {
		return err
	}

	cachedInstance, err := internalToCachedInstance(&internal)
	if err != nil {
		return err
	}

	*target = *cachedInstance
	return nil
}
