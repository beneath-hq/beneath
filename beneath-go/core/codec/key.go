package codec

import (
	"fmt"
	"time"

	"github.com/beneath-core/beneath-go/core/codec/ext/tuple"
	"github.com/beneath-core/beneath-go/core/queryparse"
)

// KeyCodec marshals data (potentially multiple values) into lexicographically sortable binary
// keys (currently used for data saved to BigTable). Preserving sort order in the binary
// encoding enables range lookups. We're using FoundationDB's encoding format, called "tuple",
// see more details in beneath-core/beneath-go/core/codec/ext/tuple
type KeyCodec struct {
	fields    []string
	avroTypes []interface{}
}

// NewKey creates a codec for encoding keys
func NewKey(fields []string, avroSchema map[string]interface{}) (*KeyCodec, error) {
	// make map of field types
	avroFields := avroSchema["fields"].([]interface{})
	avroFieldTypes := make(map[string]interface{}, len(avroFields))
	for _, avroFieldT := range avroFields {
		avroField := avroFieldT.(map[string]interface{})
		avroFieldTypes[avroField["name"].(string)] = avroField["type"]
	}

	// make list of types matching key fields
	types := make([]interface{}, len(fields))
	for idx, field := range fields {
		types[idx] = avroFieldTypes[field]
	}

	// create codec
	return &KeyCodec{
		fields:    fields,
		avroTypes: types,
	}, nil
}

// GetKeyFields returns the key fields that the codec encodes
func (c *KeyCodec) GetKeyFields() []string {
	return c.fields
}

// Marshal returns the binary representation of the key, i.e., data must
// have values for every field in fields
func (c *KeyCodec) Marshal(data map[string]interface{}) ([]byte, error) {
	// prepare tuple
	t := make(tuple.Tuple, len(c.fields))

	// add value for every keyField, preserving their order
	for idx, field := range c.fields {
		val := data[field]
		if val == nil {
			return nil, fmt.Errorf("Value for key field '%s' is nil", field)
		}

		// tuple doesn't support time -- so we encode it as an int64
		// TODO: add time handling to our fork of "tuple"
		if valTime, ok := val.(time.Time); ok {
			val = valTime.UnixNano() / int64(time.Millisecond)
		}

		t[idx] = val
	}

	// encode
	return t.Pack(), nil
}

// RangeFromQuery creates a KeyRange for the key encoded by the codec based on a query
func (c *KeyCodec) RangeFromQuery(q queryparse.Query) (kr *KeyRange, err error) {
	return NewKeyRange(q, c)
}
