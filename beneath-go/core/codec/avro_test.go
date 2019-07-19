package codec

import (
	"testing"

	"github.com/beneath-core/beneath-go/core/jsonutil"
	"github.com/beneath-core/beneath-go/core/schema"
	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"
)

func TestAvroSimple(t *testing.T) {
	avroSchema := schema.MustCompileToAvroString(`
		type Test @stream(name: "test", key: "one") {
			one: String!
			two: String!
			three: Int64!
		}
	`)

	codec, err := NewAvro(avroSchema)
	assert.Nil(t, err)

	valueJSON := []byte(`{
		"one": "testing three two one",
		"two": "0xItLooksLikeAHexButYouGotTricked",
		"three": "-9007199254741992"
	}`)

	var jsonNative, jsonNativeCopy interface{}
	err = jsonutil.UnmarshalBytes([]byte(valueJSON), &jsonNative)
	assert.Nil(t, err)
	err = jsonutil.UnmarshalBytes([]byte(valueJSON), &jsonNativeCopy)
	assert.Nil(t, err)

	binary, err := codec.Marshal(jsonNative)
	assert.Nil(t, err)

	native, err := codec.Unmarshal(binary)

	if diff := deep.Equal(jsonNativeCopy, native); diff != nil {
		t.Error(diff)
	}
}

func TestAvroComplex(t *testing.T) {
	avroSchema := schema.MustCompileToAvroString(`
		type Test @stream(name: "test", key: "one") {
			one: String!
			two: Bytes20!
			three: Int64!
			four: String!
			five: Numeric!
			six: Timestamp!
			seven: [TestA!]!
		}
		type TestA {
			one: Boolean!
			two: Float64!
		}
	`)

	codec, err := NewAvro(avroSchema)
	assert.Nil(t, err)

	valueJSON := `{
		"one": "testing one two three",
		"two": "0xaabbccddeeaabbccddeeaabbccddeeaabbccddee",
		"three": 1234567890,
		"four": "There was a bell beside the gate, and Dorothy pushed the button and heard a silvery tinkle sound within. Then the big gate swung slowly open, and they all passed through and found themselves in a high arched room, the walls of which glistened with countless emeralds. Before them stood a little man about the same size as the Munchkins. He was clothed all in green, from his head to his feet, and even his skin was of a greenish tint. At his side was a large green box.",
		"five": "-77224998599806363752588771300231266558642741460645341489178111450841839741627",
		"six": 1560949036000,
		"seven": [
			{ "one": true, "two": 3.14159265358 },
			{ "one": false, "two": 2.718281828 }
		]
	}`

	var jsonNative, jsonNativeCopy interface{}
	err = jsonutil.UnmarshalBytes([]byte(valueJSON), &jsonNative)
	assert.Nil(t, err)
	err = jsonutil.UnmarshalBytes([]byte(valueJSON), &jsonNativeCopy)
	assert.Nil(t, err)

	binary, err := codec.Marshal(jsonNative)
	assert.Nil(t, err)

	native, err := codec.Unmarshal(binary)
	assert.Nil(t, err)

	if diff := deep.Equal(jsonNativeCopy, native); diff != nil {
		t.Error(diff)
	}
}
