package codec

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/beneath-core/beneath-go/core/jsonutil"
	"github.com/beneath-core/beneath-go/core/schema"
	"github.com/go-test/deep"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

// Index represents a set of fields to generate keys for
type testIndex struct {
	indexID uuid.UUID
	fields  []string
}

func (i testIndex) GetIndexID() uuid.UUID {
	return i.indexID
}

func (i testIndex) GetFields() []string {
	return i.fields
}

func TestAvroSimple(t *testing.T) {
	index := testIndex{fields: []string{"one"}}
	avroSchema := schema.MustCompileToAvroString(`
		type Test @stream(name: "test") @key(fields: "one") {
			one: String!
			two: String!
			three: Int64!
		}
	`)

	codec, err := New(avroSchema, index, nil)
	assert.Nil(t, err)

	valueJSON := []byte(`{
		"one": "testing three two one",
		"two": "0xItLooksLikeAHexButYouGotTricked",
		"three": "-9007199254741992"
	}`)

	var jsonNative, jsonNativeCopy map[string]interface{}
	err = jsonutil.UnmarshalBytes([]byte(valueJSON), &jsonNative)
	assert.Nil(t, err)

	err = json.Unmarshal([]byte(valueJSON), &jsonNativeCopy)
	assert.Nil(t, err)

	avroNative, err := codec.ConvertToAvroNative(jsonNative, true)
	assert.Nil(t, err)

	binary, err := codec.MarshalAvro(avroNative)
	assert.Nil(t, err)

	avroNativeOut, err := codec.UnmarshalAvro(binary)
	assert.Nil(t, err)

	jsonNativeOut, err := codec.ConvertFromAvroNative(avroNativeOut, true)

	if diff := deep.Equal(jsonNativeCopy, jsonNativeOut); diff != nil {
		t.Error(diff)
	}
}

func TestAvroComplex(t *testing.T) {
	index := testIndex{fields: []string{"one"}}
	avroSchema := schema.MustCompileToAvroString(`
		type Test @stream(name: "test") @key(fields: "one") {
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
			three: Timestamp
		}
	`)

	codec, err := New(avroSchema, index, nil)
	assert.Nil(t, err)

	valueJSON := `{
		"one": "testing one two three",
		"two": "0xaabbccddeeaabbccddeeaabbccddeeaabbccddee",
		"three": 1234567890,
		"four": "There was a bell beside the gate, and Dorothy pushed the button and heard a silvery tinkle sound within. Then the big gate swung slowly open, and they all passed through and found themselves in a high arched room, the walls of which glistened with countless emeralds. Before them stood a little man about the same size as the Munchkins. He was clothed all in green, from his head to his feet, and even his skin was of a greenish tint. At his side was a large green box.",
		"five": "-77224998599806363752588771300231266558642741460645341489178111450841839741627",
		"six": 1560949036000,
		"seven": [
			{ "one": true, "two": 3.14159265358, "three": null },
			{ "one": false, "two": 2.718281828, "three": 1572445315000 }
		]
	}`

	var jsonNative, jsonNativeCopy map[string]interface{}
	err = jsonutil.UnmarshalBytes([]byte(valueJSON), &jsonNative)
	assert.Nil(t, err)

	err = json.Unmarshal([]byte(valueJSON), &jsonNativeCopy)
	assert.Nil(t, err)

	avroNative, err := codec.ConvertToAvroNative(jsonNative, true)
	assert.Nil(t, err)

	binary, err := codec.MarshalAvro(avroNative)
	assert.Nil(t, err)

	avroNativeOut, err := codec.UnmarshalAvro(binary)
	assert.Nil(t, err)

	jsonNativeOut, err := codec.ConvertFromAvroNative(avroNativeOut, true)

	if diff := deep.Equal(jsonNativeCopy, jsonNativeOut); diff != nil {
		t.Error(diff)
	}
}

func TestKeySimple(t *testing.T) {
	index := testIndex{fields: []string{"k1", "k2", "k3", "k4"}}
	schemaString := schema.MustCompileToAvroString(`
		type Test @stream(name: "test") @key(fields: ["k1", "k2", "k3", "k4"]) {
			k1: Bytes20!
			k2: Int64!
			k3: String!
			k4: Timestamp!
		}
	`)

	codec, err := New(schemaString, index, nil)
	assert.Nil(t, err)

	v1, err := codec.MarshalPrimaryKey(map[string]interface{}{
		"k1": hexToBytes("0x0000000000000000000000000000000000000000"),
		"k2": 10000000000000,
		"k3": "abc",
		"k4": time.Date(1995, time.February, 1, 0, 0, 0, 0, time.UTC),
	})
	assert.Nil(t, err)
	assert.NotNil(t, v1)

	v2, err := codec.MarshalPrimaryKey(map[string]interface{}{
		"k1": hexToBytes("0x0000000000000000000000000000000000000000"),
		"k2": 10000000000000,
		"k3": "abc",
		"k4": time.Date(1995, time.February, 2, 0, 0, 0, 0, time.UTC),
	})
	assert.Nil(t, err)
	assert.NotNil(t, v2)

	v3, err := codec.MarshalPrimaryKey(map[string]interface{}{
		"k1": hexToBytes("0x0000000000000000000000000000000000000000"),
		"k2": 10000000000000,
		"k3": "abcd",
		"k4": time.Date(1995, time.February, 1, 0, 0, 0, 0, time.UTC),
	})
	assert.Nil(t, err)
	assert.NotNil(t, v3)

	v4, err := codec.MarshalPrimaryKey(map[string]interface{}{
		"k1": hexToBytes("0x0000000000000000000000000000000000000000"),
		"k2": 90000000000000,
		"k3": "abc",
		"k4": time.Date(1995, time.February, 1, 0, 0, 0, 0, time.UTC),
	})
	assert.Nil(t, err)
	assert.NotNil(t, v4)

	v5, err := codec.MarshalPrimaryKey(map[string]interface{}{
		"k1": hexToBytes("0xFF00000000000000000000000000000000000000"),
		"k2": 10000000000000,
		"k3": "abc",
		"k4": time.Date(1995, time.February, 1, 0, 0, 0, 0, time.UTC),
	})
	assert.Nil(t, err)
	assert.NotNil(t, v5)

	_, err = codec.MarshalPrimaryKey(map[string]interface{}{
		"k1": hexToBytes("0xFF00000000000000000000000000000000000000"),
		"k2": 10000000000000,
		"k4": time.Date(1995, time.February, 1, 0, 0, 0, 0, time.UTC),
	})
	assert.NotNil(t, err)
	assert.Equal(t, "Value for index field 'k3' is nil", err.Error())

	assert.Equal(t, 60, len(v1))
	assert.Equal(t, 60, len(v2))
	assert.Equal(t, 61, len(v3))
	assert.Equal(t, 60, len(v4))
	assert.Equal(t, 59, len(v5))

	assert.Equal(t, -1, bytes.Compare(v1, v2))
	assert.Equal(t, -1, bytes.Compare(v2, v3))
	assert.Equal(t, -1, bytes.Compare(v3, v4))
	assert.Equal(t, -1, bytes.Compare(v4, v5))
}

func hexToBytes(num string) []byte {
	bytes, err := hex.DecodeString(strings.Replace(num, "0x", "", 1))
	if err != nil {
		panic(err)
	}
	return bytes
}
