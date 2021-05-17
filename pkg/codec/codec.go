package codec

import (
	"fmt"
	"sync"

	"github.com/linkedin/goavro/v2"
	uuid "github.com/satori/go.uuid"

	"github.com/beneath-hq/beneath/pkg/codec/ext/tuple"
	"github.com/beneath-hq/beneath/pkg/queryparse"
	"github.com/beneath-hq/beneath/pkg/schemalang"
	"github.com/beneath-hq/beneath/pkg/schemalang/transpilers"
)

// Index represents a set of fields to generate keys for
type Index interface {
	// GetIndexID should return a globally unique identifier for the index
	GetIndexID() uuid.UUID

	// GetShortID should return a local-to-the-stream unique identifier for the index
	GetShortID() int

	// GetFields should return the list of fields for encoding keys in the index
	GetFields() []string

	// GetNormalize should return whether indexed data should be replicated or stored by pointer
	GetNormalize() bool
}

// Codec contains schema and index info and can marshal/ummarshal records and their keys.
// Codec marshals records with Avro using the LinkedIn library (but it uses a different JSON
// representation – namely, it doesn't nest values for named/union types).
// Codec marshals keys (which may combine multiple fields) into lexicographically sortable binary
// keys (very useful for range lookups in a key-value database). We're using FoundationDB's encoding
// format, called "tuple". See more details in beneath-core/pkg/codec/ext/tuple.
type Codec struct {
	Schema           schemalang.Schema
	AvroSchema       string
	PrimaryIndex     Index
	SecondaryIndexes []Index

	avroCodec *goavro.Codec

	cacheMu           sync.Mutex
	cachedFieldTypes  map[string]*schemalang.RecordField
	cachedIndexFields map[int][]string
	cachedRefs        map[string]schemalang.Schema
}

// New creates a new Codec for encoding records between JSON and Avro and keys
func New(avroSchema string, primaryIndex Index, secondaryIndexes []Index) (*Codec, error) {
	avroCodec, err := goavro.NewCodec(avroSchema)
	if err != nil {
		return nil, fmt.Errorf("cannot create avro codec: %v", err.Error())
	}

	schema, err := transpilers.FromAvro(avroSchema)
	if err != nil {
		return nil, err
	}

	return &Codec{
		Schema:           schema,
		AvroSchema:       avroSchema,
		PrimaryIndex:     primaryIndex,
		SecondaryIndexes: secondaryIndexes,
		avroCodec:        avroCodec,
	}, nil
}

// FindIndexByShortID finds the index where GetShortID() matches shortID.
// It returns nil if no index matches.
func (c *Codec) FindIndexByShortID(shortID int) Index {
	if shortID == 0 {
		return c.PrimaryIndex
	}

	var secondary Index
	for _, index := range c.SecondaryIndexes {
		if index.GetShortID() == shortID {
			secondary = index
			break
		}
	}

	return secondary
}

// MarshalAvro returns the binary representation of the record
func (c *Codec) MarshalAvro(record map[string]interface{}) ([]byte, error) {
	avroNative, err := c.convertToAvroNative(record)
	if err != nil {
		return nil, err
	}

	binary, err := c.avroCodec.BinaryFromNative(nil, avroNative)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal record to avro: %v", err.Error())
	}

	return binary, nil
}

// UnmarshalAvro parses avro-encoded binary data to a record
func (c *Codec) UnmarshalAvro(data []byte) (map[string]interface{}, error) {
	val, remainder, err := c.avroCodec.NativeFromBinary(data)
	if err != nil {
		return nil, fmt.Errorf("couldn't unmarshal avro binary: %v", err.Error())
	} else if len(remainder) != 0 {
		return nil, fmt.Errorf("unmarshal avro binary produced remainder: data <%v> and remainder <%v>", data, remainder)
	}

	val, err = c.convertFromAvroNative(val)
	if err != nil {
		return nil, err
	}

	record, ok := val.(map[string]interface{})
	if !ok {
		panic(fmt.Errorf("Decoding avro data did not yield map"))
	}

	return record, nil
}

// ConvertFromJSONTypes converts JSON-type values in record to native Go types in accordance with the
// codec's schema.
func (c *Codec) ConvertFromJSONTypes(record map[string]interface{}) (map[string]interface{}, error) {
	val, err := jsonConverter{Refs: c.getRefs()}.convert(c.Schema, record, true)
	if err != nil {
		return nil, err
	}
	return val.(map[string]interface{}), nil
}

// ConvertToJSONTypes converts native Go types in the record to JSON-compatible types in accordance with
// the codec's schema.
func (c *Codec) ConvertToJSONTypes(record map[string]interface{}) (map[string]interface{}, error) {
	val, err := jsonConverter{Refs: c.getRefs()}.convert(c.Schema, record, false)
	if err != nil {
		return nil, err
	}
	return val.(map[string]interface{}), nil
}

// ParseIndexQuery produces a KeyRange matching one of the codec indexes
func (c *Codec) ParseIndexQuery(q queryparse.Query) (Index, KeyRange, error) {
	kr, err := newKeyRange(c, c.PrimaryIndex, q)
	if err == nil {
		return c.PrimaryIndex, kr, nil
	} else if err != ErrIndexMiss {
		return nil, kr, err
	}

	for _, index := range c.SecondaryIndexes {
		kr, err := newKeyRange(c, index, q)
		if err == nil {
			return index, kr, nil
		} else if err != ErrIndexMiss {
			return nil, kr, err
		}
	}

	return nil, kr, ErrIndexMiss
}

// MarshalKey produces a lexicographically sortable, unique binary key for the index.
// If the index is the PrimaryIndex, it's encoded outright. If the index is a SecondaryIndex,
// the primary index key is appended to ensure uniqueness.
func (c *Codec) MarshalKey(index Index, record map[string]interface{}) ([]byte, error) {
	// prepare
	fields := c.getIndexFields(index)
	t := make(tuple.Tuple, len(fields))

	// add value for every field, preserving their order
	for idx, field := range fields {
		val := record[field]
		if val == nil {
			return nil, fmt.Errorf("Value for index field '%s' is nil", field)
		}

		// convert UUID because package tuple has its own UUID type
		if uuidVal, ok := val.(uuid.UUID); ok {
			val = tuple.UUID(uuidVal)
		}

		t[idx] = val
	}

	// encode
	return t.Pack(), nil
}

// UnmarshalKey decodes a key marshalled with MarshalKey.
// If index is a secondary index, both secondary and primary fields are included in the result.
func (c *Codec) UnmarshalKey(index Index, key []byte) (map[string]interface{}, error) {
	fields := c.getIndexFields(index)

	t, err := tuple.Unpack(key)
	if err != nil {
		return nil, err
	}

	if len(t) != len(fields) {
		return nil, fmt.Errorf("error in UnmarshalKey: values in key doesn't match number of fields expected for the index")
	}

	result := make(map[string]interface{}, len(t))
	for i, val := range t {
		result[fields[i]] = val
	}

	return result, nil
}

func (c *Codec) getFieldTypes() map[string]*schemalang.RecordField {
	if c.cachedFieldTypes != nil {
		return c.cachedFieldTypes
	}

	c.cacheMu.Lock()
	if c.cachedFieldTypes == nil {
		recordSchema := c.Schema.(*schemalang.Record)
		c.cachedFieldTypes = make(map[string]*schemalang.RecordField, len(recordSchema.Fields))
		for _, field := range recordSchema.Fields {
			c.cachedFieldTypes[field.Name] = field
		}
	}
	c.cacheMu.Unlock()

	return c.cachedFieldTypes
}

func (c *Codec) getIndexFields(index Index) []string {
	// secondary index keys are suffixed with the primary key fields, but excluding duplicate fields.
	// this function computes the de-duplicated list of fields for an index and memoizes the results

	shortID := index.GetShortID()
	primary := index.GetShortID() == 0

	if primary {
		return index.GetFields()
	}

	c.cacheMu.Lock()

	if c.cachedIndexFields == nil {
		c.cachedIndexFields = make(map[int][]string)
	}

	fields := c.cachedIndexFields[shortID]
	if len(fields) == 0 {
		fields = index.GetFields()
		for _, primaryField := range c.PrimaryIndex.GetFields() {
			add := true
			for _, field := range fields {
				if field == primaryField {
					add = false
					break
				}
			}
			if add {
				fields = append(fields, primaryField)
			}
		}
		c.cachedIndexFields[shortID] = fields
	}

	c.cacheMu.Unlock()

	return fields
}

func (c *Codec) getRefs() map[string]schemalang.Schema {
	if c.cachedRefs != nil {
		return c.cachedRefs
	}
	c.cacheMu.Lock()
	if c.cachedRefs == nil {
		c.cachedRefs = schemalang.ExtractRefs(c.Schema)
	}
	c.cacheMu.Unlock()
	return c.cachedRefs
}
