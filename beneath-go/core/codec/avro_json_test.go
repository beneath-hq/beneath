package codec

import (
	"encoding/json"
	"testing"

	"github.com/beneath-core/beneath-go/core/jsonutil"
	"github.com/beneath-core/beneath-go/core/schema"

	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"
)

func TestAvroJson1(t *testing.T) {
	avroSchema := schema.MustCompileToAvro(`
		type Test @stream(name: "test", key: ["one", "two"]) {
			one: String!
			two: Timestamp!
			three: TestA!
		}
		type TestA {
			one: Bytes!
			two: Numeric!
			three: Int
			four: [Int!]
			five: [Bytes10!]
			six: [TestB!]
			seven: TestB
		}
		type TestB {
			one: Int!
			two: Bytes10
		}
	`)

	valueJSON := `{
		"one": "aaa bbb ccc",
		"two": 1561294281000,
		"three": {
			"one": "0xbeef",
			"two": "12345678901234567890123456789012345678901234567890",
			"three": null,
			"four": [
				100,
				200
			],
			"five": [
				"0x00112233445566778899",
				"0x99887766554433221100"
			],
			"six": null,
			"seven": {
				"one": 31,
				"two": "0x99887766554433221100"
			}
		}
	}`

	var value, valueCopy interface{}
	err := jsonutil.UnmarshalBytes([]byte(valueJSON), &value)
	assert.Nil(t, err)

	// unmarshal valueCopy with regular json to avoid getting json.Number values
	err = json.Unmarshal([]byte(valueJSON), &valueCopy)
	assert.Nil(t, err)

	avroNative, err := jsonNativeToAvroNative(avroSchema, value, map[string]interface{}{})
	assert.Nil(t, err)
	jsonNative, err := avroNativeToJSONNative(avroSchema, avroNative, map[string]interface{}{}, true)
	assert.Nil(t, err)

	if diff := deep.Equal(jsonNative, valueCopy); diff != nil {
		t.Error(diff)
	}
}
