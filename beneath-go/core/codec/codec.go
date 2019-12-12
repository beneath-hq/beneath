package codec

import (
	"encoding/json"
	"fmt"

	"github.com/linkedin/goavro/v2"
	uuid "github.com/satori/go.uuid"

	"github.com/beneath-core/beneath-go/core/codec/ext/tuple"
	"github.com/beneath-core/beneath-go/core/queryparse"
)

// Codec marshals keys (potentially multiple values) into lexicographically sortable binary
// keys (currently used for data saved to BigTable). Preserving sort order in the binary
// encoding enables range lookups. We're using FoundationDB's encoding format, called "tuple",
// see more details in beneath-core/beneath-go/core/codec/ext/tuple

// Codec marshals records with avro, but we can't use the LinkedIn library directly because
// we're representing data as JSON in a different way

// Codec contains schema info for a stream and can marshal/ummarshal records and their keys
type Codec struct {
	avroCodec        *goavro.Codec
	avroSchema       map[string]interface{}
	avroSchemaString string
	avroFieldTypes   map[string]interface{}
	primaryIndex     Index
	secondaryIndexes []Index
}

// Index represents a set of fields to generate keys for
type Index interface {
	// GetIndexID should return a globally unique identifier for the index
	GetIndexID() uuid.UUID

	// GetFields should return the list of fields for encoding keys in the index
	GetFields() []string
}

// New creates a new Codec for encoding records between JSON and Avro and keys
func New(avroSchema string, primaryIndex Index, secondaryIndexes []Index) (*Codec, error) {
	codec := &Codec{}
	codec.avroSchemaString = avroSchema
	codec.primaryIndex = primaryIndex
	codec.secondaryIndexes = secondaryIndexes

	// parse avro schema
	err := json.Unmarshal([]byte(avroSchema), &codec.avroSchema)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal schema: %v", err.Error())
	}

	// create avro codec
	codec.avroCodec, err = goavro.NewCodec(avroSchema)
	if err != nil {
		return nil, fmt.Errorf("cannot create avro codec: %v", err.Error())
	}

	// make map of field types (for efficient lookup when encoding keys)
	avroFields := codec.avroSchema["fields"].([]interface{})
	avroFieldTypes := make(map[string]interface{}, len(avroFields))
	for _, avroFieldT := range avroFields {
		avroField := avroFieldT.(map[string]interface{})
		avroFieldTypes[avroField["name"].(string)] = avroField["type"]
	}
	codec.avroFieldTypes = avroFieldTypes

	return codec, nil
}

// GetPrimaryIndex returns the primary index passed on creation
func (c *Codec) GetPrimaryIndex() Index {
	return c.primaryIndex
}

// GetSecondaryIndexes returns the secondary indexes passed on creation
func (c *Codec) GetSecondaryIndexes() []Index {
	return c.secondaryIndexes
}

// GetAvroSchema returns the avro schema as a map
func (c *Codec) GetAvroSchema() map[string]interface{} {
	return c.avroSchema
}

// GetAvroSchemaString returns the avro schema as a string
func (c *Codec) GetAvroSchemaString() string {
	return c.avroSchemaString
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

	obj, err := jsonNativeToAvroNative(c.avroSchema, data, map[string]interface{}{})
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
	obj, err := avroNativeToJSONNative(c.avroSchema, avroNative, map[string]interface{}{}, convertToJSONTypes)
	if err != nil {
		return nil, err
	}

	res, ok := obj.(map[string]interface{})
	if !ok {
		panic(fmt.Errorf("Output type of avroNativeToJSONNative does not match input type"))
	}

	return res, nil
}

// MarshalPrimaryKey returns the binary representation of the record key.
// Input must have been parsed into Avro native form
func (c *Codec) MarshalPrimaryKey(data map[string]interface{}) ([]byte, error) {
	return c.marshalKey(data, c.primaryIndex)
}

func (c *Codec) marshalKey(data map[string]interface{}, index Index) ([]byte, error) {
	// prepare tuple
	t := make(tuple.Tuple, len(index.GetFields()))

	// add value for every keyField, preserving their order
	for idx, field := range index.GetFields() {
		val := data[field]
		if val == nil {
			return nil, fmt.Errorf("Value for index field '%s' is nil", field)
		}

		t[idx] = val
	}

	// encode
	return t.Pack(), nil
}

// MakeKeyRange creates a KeyRange for the key encoded by the codec based on a query and a page offset (after)
func (c *Codec) MakeKeyRange(where queryparse.Query, after queryparse.Query) (KeyRange, error) {
	kr, err := NewKeyRange(c, where)
	if err != nil {
		return kr, err
	}

	if after != nil {
		kr, err = kr.WithAfter(c, after)
		if err != nil {
			return kr, err
		}
	}

	return kr, nil
}
