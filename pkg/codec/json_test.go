package codec

import (
	"encoding/json"
	"testing"

	"github.com/beneath-hq/beneath/pkg/jsonutil"
	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"
)

func TestAvroJson1(t *testing.T) {
	avroSchemaString := transpileGraphQLToAvroString(`
		type Test @schema(name: "test") @key(fields: ["one", "two"]) {
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
			three: Float64
			four: Float64
			five: Float64
		}
	`)

	// 0xbeef = vu8=
	// 0x00112233445566778899 = ABEiM0RVZneImQ==
	// 0x99887766554433221100 = mYh3ZlVEMyIRAA==

	valueJSON := `{
		"one": "aaa bbb ccc",
		"two": 1561294281000,
		"three": {
			"one": "vu8=",
			"two": "12345678901234567890123456789012345678901234567890",
			"three": null,
			"four": [
				100,
				200
			],
			"five": [
				"ABEiM0RVZneImQ==",
				"mYh3ZlVEMyIRAA=="
			],
			"six": null,
			"seven": {
				"one": 31,
				"two": "mYh3ZlVEMyIRAA==",
				"three": "NaN",
				"four": "-Infinity",
				"five": 3.141
			}
		}
	}`

	var value, valueCopy map[string]interface{}
	err := jsonutil.UnmarshalBytes([]byte(valueJSON), &value)
	assert.Nil(t, err)

	// unmarshal valueCopy with regular json to avoid getting json.Number values
	err = json.Unmarshal([]byte(valueJSON), &valueCopy)
	assert.Nil(t, err)

	codec, err := New(avroSchemaString, nil, nil)
	assert.Nil(t, err)

	avroNative, err := codec.ConvertFromJSONTypes(value)
	assert.Nil(t, err)
	jsonNative, err := codec.ConvertToJSONTypes(avroNative)
	assert.Nil(t, err)

	if diff := deep.Equal(jsonNative, valueCopy); diff != nil {
		t.Error(diff)
	}
}
