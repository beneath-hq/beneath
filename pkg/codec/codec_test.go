package codec

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"gitlab.com/beneath-org/beneath/pkg/codec/ext/tuple"

	"gitlab.com/beneath-org/beneath/pkg/queryparse"

	"github.com/go-test/deep"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"

	"gitlab.com/beneath-org/beneath/pkg/jsonutil"
	"gitlab.com/beneath-org/beneath/pkg/schema"
)

// Index represents a set of fields to generate keys for
type testIndex struct {
	indexID uuid.UUID
	fields  []string
	shortID int
}

func (i testIndex) GetIndexID() uuid.UUID {
	return i.indexID
}

func (i testIndex) GetShortID() int {
	return i.shortID
}

func (i testIndex) GetFields() []string {
	return i.fields
}

func (i testIndex) GetNormalize() bool {
	return false
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

	v1, err := codec.MarshalKey(codec.PrimaryIndex, map[string]interface{}{
		"k1": hexToBytes("0x0000000000000000000000000000000000000000"),
		"k2": 10000000000000,
		"k3": "abc",
		"k4": time.Date(1995, time.February, 1, 0, 0, 0, 0, time.UTC),
	})
	assert.Nil(t, err)
	assert.NotNil(t, v1)

	v2, err := codec.MarshalKey(codec.PrimaryIndex, map[string]interface{}{
		"k1": hexToBytes("0x0000000000000000000000000000000000000000"),
		"k2": 10000000000000,
		"k3": "abc",
		"k4": time.Date(1995, time.February, 2, 0, 0, 0, 0, time.UTC),
	})
	assert.Nil(t, err)
	assert.NotNil(t, v2)

	v3, err := codec.MarshalKey(codec.PrimaryIndex, map[string]interface{}{
		"k1": hexToBytes("0x0000000000000000000000000000000000000000"),
		"k2": 10000000000000,
		"k3": "abcd",
		"k4": time.Date(1995, time.February, 1, 0, 0, 0, 0, time.UTC),
	})
	assert.Nil(t, err)
	assert.NotNil(t, v3)

	v4, err := codec.MarshalKey(codec.PrimaryIndex, map[string]interface{}{
		"k1": hexToBytes("0x0000000000000000000000000000000000000000"),
		"k2": 90000000000000,
		"k3": "abc",
		"k4": time.Date(1995, time.February, 1, 0, 0, 0, 0, time.UTC),
	})
	assert.Nil(t, err)
	assert.NotNil(t, v4)

	v5, err := codec.MarshalKey(codec.PrimaryIndex, map[string]interface{}{
		"k1": hexToBytes("0xFF00000000000000000000000000000000000000"),
		"k2": 10000000000000,
		"k3": "abc",
		"k4": time.Date(1995, time.February, 1, 0, 0, 0, 0, time.UTC),
	})
	assert.Nil(t, err)
	assert.NotNil(t, v5)

	_, err = codec.MarshalKey(codec.PrimaryIndex, map[string]interface{}{
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

func TestSecondaryKeys(t *testing.T) {
	primary := testIndex{
		fields:  []string{"k1", "k2"},
		shortID: 0}
	secondary := testIndex{
		fields:  []string{"k2", "k3"},
		shortID: 1}
	schemaString := schema.MustCompileToAvroString(`
		type Test @stream(name: "test") @key(fields: ["k1", "k2"]) {
			k1: Bytes20!
			k2: Int64!
			k3: String!
			k4: Timestamp!
		}
	`)

	codec, err := New(schemaString, primary, []Index{secondary})
	assert.Nil(t, err)

	m1 := map[string]interface{}{
		"k1": hexToBytes("0x0000000000000000000000000000000000000000"),
		"k2": int64(10000000000000),
	}
	v1, err := codec.MarshalKey(codec.PrimaryIndex, m1)
	assert.Nil(t, err)
	assert.NotNil(t, v1)
	s1, err := codec.UnmarshalKey(codec.PrimaryIndex, v1)
	assert.Nil(t, err)
	assert.NotNil(t, s1)
	assert.True(t, reflect.DeepEqual(m1, s1))

	m2 := map[string]interface{}{
		"k1": hexToBytes("0x0000000000000000000000000000000000000000"),
		"k2": int64(10000000000000),
		"k3": "hello",
	}
	v2, err := codec.MarshalKey(codec.SecondaryIndexes[0], m2)
	assert.Nil(t, err)
	assert.NotNil(t, v2)
	assert.True(t, bytes.Equal(v2, tuple.Tuple{m2["k2"], m2["k3"], m2["k1"]}.Pack()))
	s2, err := codec.UnmarshalKey(codec.SecondaryIndexes[0], v2)
	assert.Nil(t, err)
	assert.NotNil(t, s2)
	assert.True(t, reflect.DeepEqual(m2, s2))
}

func TestQueryParse(t *testing.T) {
	index1 := testIndex{fields: []string{"one"}}
	index2 := testIndex{fields: []string{"two"}}
	index3 := testIndex{fields: []string{"three", "two"}}
	avroSchema := schema.MustCompileToAvroString(`
		type Test @stream(name: "test") @key(fields: "one") @index(fields: "two") @index(fields: ["three", "two"]) {
			one: String!
			two: Bytes!
			three: Int64!
			four: Int64!
		}
	`)

	codec, err := New(avroSchema, index1, []Index{index2, index3})
	assert.Nil(t, err)

	q1, err := queryparse.JSONStringToQuery("")
	assert.Nil(t, err)
	idx1, kr1, err := codec.ParseQuery(q1)
	assert.Nil(t, err)
	assert.Equal(t, index1, idx1)
	assert.True(t, bytes.Equal([]byte{}, kr1.Base))

	q2, err := queryparse.JSONStringToQuery(`{"two":{"_prefix": "0xAAAA"}}`)
	assert.Nil(t, err)
	idx2, kr2, err := codec.ParseQuery(q2)
	assert.Nil(t, err)
	assert.Equal(t, index2, idx2)
	assert.True(t, kr2.Contains(tuple.Tuple{[]byte{0xAA, 0xAA, 0xBB}}.Pack()))
	assert.False(t, kr2.Contains(tuple.Tuple{[]byte{0xAA, 0xBB}}.Pack()))

	q3, err := queryparse.JSONStringToQuery(`{"three": 1000, "two": {"_prefix": "0xAAAA"}}`)
	assert.Nil(t, err)
	idx3, kr3, err := codec.ParseQuery(q3)
	assert.Nil(t, err)
	assert.Equal(t, index3, idx3)
	assert.True(t, kr3.Contains(tuple.Tuple{1000, []byte{0xAA, 0xAA, 0xBB}}.Pack()))
	assert.False(t, kr3.Contains(tuple.Tuple{1000, []byte{0xAA, 0xBB}}.Pack()))

	q4, err := queryparse.JSONStringToQuery(`{"four": 1000}`)
	assert.Nil(t, err)
	_, _, err = codec.ParseQuery(q4)
	assert.NotNil(t, err)
	assert.Equal(t, ErrIndexMiss, err)
}

func hexToBytes(num string) []byte {
	bytes, err := hex.DecodeString(strings.Replace(num, "0x", "", 1))
	if err != nil {
		panic(err)
	}
	return bytes
}
