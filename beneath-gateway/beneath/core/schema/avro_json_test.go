package schema

import (
	"encoding/json"
	"testing"

	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"
)

func TestAvroJson1(t *testing.T) {
	var schema interface{}
	err := json.Unmarshal([]byte(`{
		"name": "test",
		"type": "record",
		"indexes": [["key_one", "key_two"]],
		"fields": [
			{"name": "key_one", "type": "string"},
			{"name": "key_two", "type": "long", "logicalType": "timestamp-millis"},
			{"name": "value_one", "type": "record", "fields": [
				{"name": "one", "type": "bytes"},
				{"name": "two", "type": "bytes", "logicalType": "decimal"},
				{"name": "three", "type": ["null", "int"]},
				{"name": "four", "type": "array", "items": [
					"null",
					{"type": "fixed", "size": 10}
				]},
				{"name": "five", "type": [
					"null",
					{"name": "four_type", "type": "record", "fields": [
						{"type": "int", "name": "four_one"}
					]}
				]}
			]}
		]
	}`), &schema)
	assert.Nil(t, err)

	valueJSON := `{
		"key_one": "aaa bbb ccc",
		"key_two": 1561294281000,
		"value_one": {
			"one": "0xbeef",
			"two": "12345678901234567890123456789012345678901234567890",
			"three": null,
			"four": [
				"0x00112233445566778899",
				null,
				"0x99887766554433221100"
			],
			"five": {
				"four_one": 31
			}
		}
	}`

	var value, valueCopy interface{}
	err = json.Unmarshal([]byte(valueJSON), &value)
	assert.Nil(t, err)
	err = json.Unmarshal([]byte(valueJSON), &valueCopy)
	assert.Nil(t, err)

	avroNative, err := jsonNativeToAvroNative(schema, value)
	assert.Nil(t, err)
	jsonNative, err := avroNativeToJSONNative(schema, avroNative)
	assert.Nil(t, err)

	if diff := deep.Equal(jsonNative, valueCopy); diff != nil {
		t.Error(diff)
	}
}
