package codec

import (
	"encoding/json"
	"fmt"

	"github.com/linkedin/goavro/v2"
)

// AvroCodec contains schema info for a stream and can marshal/ummarshal records
type AvroCodec struct {
	avroSchemaString string
	avroSchema       map[string]interface{}
	avroCodec        *goavro.Codec
}

// NewAvro creates a new Codec for encoding data between JSON and Avro
// We're not just using the LinkedIn version because we want to handle
// JSON data more naturally (and more opinionated, to be fair)
func NewAvro(avroSchema string) (*AvroCodec, error) {
	codec := &AvroCodec{}
	codec.avroSchemaString = avroSchema

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

	return codec, nil
}

// GetSchema returns the avro schema as a map
func (c *AvroCodec) GetSchema() map[string]interface{} {
	return c.avroSchema
}

// GetSchemaString returns the avro schema as a string
func (c *AvroCodec) GetSchemaString() string {
	return c.avroSchemaString
}

// Marshal maps an unmarshaled json object to avro-encoded binary
func (c *AvroCodec) Marshal(jsonNative interface{}) ([]byte, error) {
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
func (c *AvroCodec) Unmarshal(data []byte) (interface{}, error) {
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
