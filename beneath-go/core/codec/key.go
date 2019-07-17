package codec

import (
	"fmt"
	"time"

	"github.com/beneath-core/beneath-go/core/codec/ext/tuple"
)

// KeyCodec marshals data (potentially multiple values) into lexicographically sortable binary
// keys (currently used for data saved to BigTable). Preserving sort order in the binary
// encoding enables range lookups. We're using FoundationDB's encoding format, called "tuple",
// see more details in beneath-core/beneath-go/core/codec/ext/tuple
type KeyCodec struct {
	keyFields []string
}

// NewKey creates a codec for encoding keys
func NewKey(keyFields []string, avroSchema map[string]interface{}) (*KeyCodec, error) {
	codec := &KeyCodec{
		keyFields: keyFields,
	}
	return codec, nil
}

// GetKeyFields returns the key fields that the codec encodes
func (c *KeyCodec) GetKeyFields() []string {
	return c.keyFields
}

// Marshal returns the binary representation of the key, i.e., data must
// have values for every field in keyFields
func (c *KeyCodec) Marshal(data map[string]interface{}) ([]byte, error) {
	// prepare tuple
	t := make(tuple.Tuple, len(c.keyFields))

	// add value for every keyField, preserving their order
	for idx, field := range c.keyFields {
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
