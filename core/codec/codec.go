package codec

import (
	"encoding/json"
	"fmt"

	"github.com/linkedin/goavro/v2"
	uuid "github.com/satori/go.uuid"

	"github.com/beneath-core/core/codec/ext/tuple"
	"github.com/beneath-core/core/queryparse"
)

// Codec marshals keys (potentially multiple values) into lexicographically sortable binary
// keys (currently used for data saved to BigTable). Preserving sort order in the binary
// encoding enables range lookups. We're using FoundationDB's encoding format, called "tuple",
// see more details in beneath-core/core/codec/ext/tuple

// Codec marshals records with avro, but we can't use the LinkedIn library directly because
// we're representing data as JSON in a different way

// Codec contains schema info for a stream and can marshal/ummarshal records and their keys
type Codec struct {
	AvroSchema       map[string]interface{}
	AvroSchemaString string
	PrimaryIndex     Index
	SecondaryIndexes []Index

	avroCodec        *goavro.Codec
	avroFieldTypes   map[string]interface{}
	indexFieldsCache map[int][]string
}

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

// New creates a new Codec for encoding records between JSON and Avro and keys
func New(avroSchema string, primaryIndex Index, secondaryIndexes []Index) (*Codec, error) {
	codec := &Codec{}
	codec.AvroSchemaString = avroSchema
	codec.PrimaryIndex = primaryIndex
	codec.SecondaryIndexes = secondaryIndexes
	codec.indexFieldsCache = make(map[int][]string)

	// parse avro schema
	err := json.Unmarshal([]byte(avroSchema), &codec.AvroSchema)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal schema: %v", err.Error())
	}

	// create avro codec
	codec.avroCodec, err = goavro.NewCodec(avroSchema)
	if err != nil {
		return nil, fmt.Errorf("cannot create avro codec: %v", err.Error())
	}

	// make map of field types (for efficient lookup when encoding keys)
	avroFields := codec.AvroSchema["fields"].([]interface{})
	avroFieldTypes := make(map[string]interface{}, len(avroFields))
	for _, avroFieldT := range avroFields {
		avroField := avroFieldT.(map[string]interface{})
		avroFieldTypes[avroField["name"].(string)] = avroField["type"]
	}
	codec.avroFieldTypes = avroFieldTypes

	return codec, nil
}

// MarshalAvro returns the binary representation of the record
// Input must have been parsed into Avro native form
func (c *Codec) MarshalAvro(avroNative map[string]interface{}) ([]byte, error) {
	binary, err := c.avroCodec.BinaryFromNative(nil, avroNative)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal record to avro: %v", err.Error())
	}

	return binary, nil
}

// UnmarshalAvro maps avro-encoded binary to a map in Avro native form
func (c *Codec) UnmarshalAvro(data []byte) (map[string]interface{}, error) {
	obj, remainder, err := c.avroCodec.NativeFromBinary(data)
	if err != nil {
		return nil, fmt.Errorf("couldn't unmarshal avro binary: %v", err.Error())
	} else if len(remainder) != 0 {
		return nil, fmt.Errorf("unmarshal avro binary produced remainder: data <%v> and remainder <%v>", data, remainder)
	}

	res, ok := obj.(map[string]interface{})
	if !ok {
		panic(fmt.Errorf("Decoding avro data did not yield map"))
	}

	return res, nil
}

// ConvertToAvroNative converts a record (possibly decoded from JSON) to avro native representation
func (c *Codec) ConvertToAvroNative(data map[string]interface{}, convertFromJSONTypes bool) (map[string]interface{}, error) {
	if !convertFromJSONTypes {
		panic(fmt.Errorf("ConvertToAvroNative currently only supports convertFromJSONTypes == true"))
	}

	obj, err := jsonNativeToAvroNative(c.AvroSchema, data, map[string]interface{}{})
	if err != nil {
		return nil, err
	}

	res, ok := obj.(map[string]interface{})
	if !ok {
		panic(fmt.Errorf("Output type of jsonNativeToAvroNative does not match input type"))
	}

	return res, nil
}

// ConvertFromAvroNative is the reverse operation of ConvertToAvroNative with an option
// to convert types to JSON friendly types
func (c *Codec) ConvertFromAvroNative(avroNative map[string]interface{}, convertToJSONTypes bool) (map[string]interface{}, error) {
	obj, err := avroNativeToJSONNative(c.AvroSchema, avroNative, map[string]interface{}{}, convertToJSONTypes)
	if err != nil {
		return nil, err
	}

	res, ok := obj.(map[string]interface{})
	if !ok {
		panic(fmt.Errorf("Output type of avroNativeToJSONNative does not match input type"))
	}

	return res, nil
}

// MarshalKey produces a lexicographically sortable, unique binary key for the index.
// If the index is the PrimaryIndex, it's encoded outright.
// If the index is a SecondaryIndex, the primary index key is appended to ensure uniqueness.
func (c *Codec) MarshalKey(index Index, data map[string]interface{}) ([]byte, error) {
	// prepare
	fields := c.getIndexFields(index)
	t := make(tuple.Tuple, len(fields))

	// add value for every field, preserving their order
	for idx, field := range fields {
		val := data[field]
		if val == nil {
			return nil, fmt.Errorf("Value for index field '%s' is nil", field)
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

// secondary index keys are suffixed with the primary key fields, but excluding duplicate fields.
// this function computes the de-duplicated list of fields for an index and memoizes the results
func (c *Codec) getIndexFields(index Index) []string {
	shortID := index.GetShortID()
	primary := index.GetShortID() == 0

	if primary {
		return index.GetFields()
	}

	fields := c.indexFieldsCache[shortID]
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

		c.indexFieldsCache[shortID] = fields
	}

	return fields
}

// ParseQuery produces a KeyRange matching one of the codec indexes
func (c *Codec) ParseQuery(q queryparse.Query) (Index, KeyRange, error) {
	kr, err := NewKeyRange(c, c.PrimaryIndex, q)
	if err == nil {
		return c.PrimaryIndex, kr, nil
	} else if err != ErrIndexMiss {
		return nil, kr, err
	}

	for _, index := range c.SecondaryIndexes {
		kr, err := NewKeyRange(c, index, q)
		if err == nil {
			return index, kr, nil
		} else if err != ErrIndexMiss {
			return nil, kr, err
		}
	}

	return nil, kr, ErrIndexMiss
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
