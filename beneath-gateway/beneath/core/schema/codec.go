package schema

import (
	"encoding/json"
	"fmt"

	"github.com/linkedin/goavro/v2"
)

// Codec contains schema info for a stream and can marshal/ummarshal records
type Codec struct {
	indexes    [][]string
	avroSchema map[string]interface{}
	avroCodec  *goavro.Codec
}

// NewCodec creates a new Codec
func NewCodec(schemaJSON string) (*Codec, error) {
	codec := &Codec{}

	err := json.Unmarshal([]byte(schemaJSON), &codec.avroSchema)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal schema: %v", err.Error())
	}

	ValidateSchema(codec.avroSchema)

	indexes := codec.avroSchema["indexes"].([][]string)
	codec.indexes = indexes
	delete(codec.avroSchema, "indexes")

	codec.avroCodec, err = goavro.NewCodec(schemaJSON)
	if err != nil {
		return nil, fmt.Errorf("cannot create avro codec: %v", err.Error())
	}

	return codec, nil
}

// Marshal maps an unmarshaled json object to avro-encoded binary
func (c *Codec) Marshal(jsonNative interface{}) ([]byte, error) {
	avroNative, err := jsonNativeToAvroNative(c.avroSchema, jsonNative)
	if err != nil {
		return nil, err
	}

	binary, err := c.avroCodec.BinaryFromNative(nil, avroNative)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal record to avro: %v", err.Error())
	}

	return binary, nil
}

// Unmarshal maps avro-encoded binary to an object that can be marshaled to json
func (c *Codec) Unmarshal(data []byte) (interface{}, error) {
	avroNative, remainder, err := c.avroCodec.NativeFromBinary(data)
	if err != nil {
		return nil, fmt.Errorf("couldn't unmarshal avro binary: %v", err.Error())
	} else if len(remainder) != 0 {
		return nil, fmt.Errorf("unmarshal avro binary produced remainder: data <%v> and remainder <%v>", data, remainder)
	}

	jsonNative, err := avroNativeToJSONNative(c.avroSchema, avroNative)
	if err != nil {
		return nil, err
	}

	return jsonNative, nil
}
