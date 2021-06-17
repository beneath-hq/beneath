package codec

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/go-test/deep"
	"github.com/linkedin/goavro/v2"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"

	"github.com/beneath-hq/beneath/pkg/codec/ext/tuple"
	"github.com/beneath-hq/beneath/pkg/jsonutil"
	"github.com/beneath-hq/beneath/pkg/queryparse"
	"github.com/beneath-hq/beneath/pkg/schemalang"
	"github.com/beneath-hq/beneath/pkg/schemalang/transpilers"
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

func transpileGraphQLToAvroString(gql string) string {
	s, _, err := transpilers.FromGraphQL(gql)
	if err != nil {
		panic(err)
	}
	err = schemalang.Check(s)
	if err != nil {
		panic(err)
	}
	return transpilers.ToAvro(s, false)
}

func TestAvroSimple(t *testing.T) {
	index := testIndex{fields: []string{"one"}}
	avroSchema := transpileGraphQLToAvroString(`
		type Test @schema(name: "test") @key(fields: "one") {
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

	native, err := codec.ConvertFromJSONTypes(jsonNative)
	assert.Nil(t, err)

	binary, err := codec.MarshalAvro(native)
	assert.Nil(t, err)

	nativeOut, err := codec.UnmarshalAvro(binary)
	assert.Nil(t, err)

	jsonNativeOut, err := codec.ConvertToJSONTypes(nativeOut)

	if diff := deep.Equal(jsonNativeCopy, jsonNativeOut); diff != nil {
		t.Error(diff)
	}
}

func TestAvroComplex(t *testing.T) {
	index := testIndex{fields: []string{"one"}}
	avroSchema := transpileGraphQLToAvroString(`
		type Test @schema(name: "test") @key(fields: "one") {
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

	// 0xaabbccddeeaabbccddeeaabbccddeeaabbccddee = qrvM3e6qu8zd7qq7zN3uqrvM3e4=

	valueJSON := `{
		"one": "testing one two three",
		"two": "qrvM3e6qu8zd7qq7zN3uqrvM3e4=",
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

	native, err := codec.ConvertFromJSONTypes(jsonNative)
	assert.Nil(t, err)

	binary, err := codec.MarshalAvro(native)
	assert.Nil(t, err)

	nativeOut, err := codec.UnmarshalAvro(binary)
	assert.Nil(t, err)

	jsonNativeOut, err := codec.ConvertToJSONTypes(nativeOut)

	if diff := deep.Equal(jsonNativeCopy, jsonNativeOut); diff != nil {
		t.Error(diff)
	}
}

func TestAvroVeryComplex(t *testing.T) {
	avroSchema := transpileGraphQLToAvroString(`
		" Docs, docs, docs, docs! "
		type Test @schema @key(fields: ["a", "b"]) {
			a: String!
			b: Timestamp!
			c: Bytes20
			d: TestB
			e: [TestB!]!
		}
		type TestB {
			a: Boolean!
			b: Bytes!
			c: Bytes20
			d: Float!
			e: Float32!
			f: Int!
			g: Int64!
			h: Numeric!
			i: String!
			j: Timestamp!
			k: TestC
			l: [TestC!]
		}
		enum TestC {
			Aa
			Bb
			Cc
		}
	`)

	codec, err := goavro.NewCodec(avroSchema)
	assert.Nil(t, err)
	assert.NotNil(t, codec)

	json1 := `{
		"a": "Example example",
		"b": 1562270699000,
		"c": { "bytes20": "aaaaaaaaaaaaaaaaaaaa" },
		"d": null,
		"e": [
			{
				"a": true,
				"b": "sfjhewuiqbcu fjoidshfoew",
				"c": null,
				"d": 98765432450.34567890765432,
				"e": 123.76432,
				"f": 1844674000000000000,
				"g": 1844674000000000001,
				"h": "999999999999999999999999999999999999999999999999999999999999999999999999",
				"i": "hello world",
				"j": 1562270699000,
				"k": {"TestC": "Bb"},
				"l": {"array": ["Cc", "Aa"]}
			}
		]
	}`

	native1, _, err := codec.NativeFromTextual([]byte(json1))
	assert.Nil(t, err)

	// hack 1
	n, _ := big.NewRat(1, 1).SetString("999999999999999999999999999999999999999999999999999999999999999999999999")
	native1.(map[string]interface{})["e"].([]interface{})[0].(map[string]interface{})["h"] = n

	binary, err := codec.BinaryFromNative(nil, native1)
	assert.Nil(t, err)

	native2, _, err := codec.NativeFromBinary(binary)
	assert.Nil(t, err)

	// hack 2
	native2.(map[string]interface{})["e"].([]interface{})[0].(map[string]interface{})["h"] = n

	assert.Equal(t, reflect.DeepEqual(native1, native2), true)
}

func TestKeySimple(t *testing.T) {
	index := testIndex{fields: []string{"k1", "k2", "k3", "k4"}}
	schemaString := transpileGraphQLToAvroString(`
		type Test @schema(name: "test") @key(fields: ["k1", "k2", "k3", "k4"]) {
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
	schemaString := transpileGraphQLToAvroString(`
		type Test @schema(name: "test") @key(fields: ["k1", "k2"]) {
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
	avroSchema := transpileGraphQLToAvroString(`
		type Test @schema(name: "test") @key(fields: "one") @index(fields: "two") @index(fields: ["three", "two"]) {
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
	idx1, kr1, err := codec.ParseIndexQuery(q1)
	assert.Nil(t, err)
	assert.Equal(t, index1, idx1)
	assert.True(t, bytes.Equal([]byte{}, kr1.Base))

	// 0xAAAA = qqo=

	q2, err := queryparse.JSONStringToQuery(`{"two":{"_prefix": "qqo="}}`)
	assert.Nil(t, err)
	idx2, kr2, err := codec.ParseIndexQuery(q2)
	assert.Nil(t, err)
	assert.Equal(t, index2, idx2)
	assert.True(t, kr2.Contains(tuple.Tuple{[]byte{0xAA, 0xAA, 0xBB}}.Pack()))
	assert.False(t, kr2.Contains(tuple.Tuple{[]byte{0xAA, 0xBB}}.Pack()))

	q3, err := queryparse.JSONStringToQuery(`{"three": 1000, "two": {"_prefix": "qqo="}}`)
	assert.Nil(t, err)
	idx3, kr3, err := codec.ParseIndexQuery(q3)
	assert.Nil(t, err)
	assert.Equal(t, index3, idx3)
	assert.True(t, kr3.Contains(tuple.Tuple{1000, []byte{0xAA, 0xAA, 0xBB}}.Pack()))
	assert.False(t, kr3.Contains(tuple.Tuple{1000, []byte{0xAA, 0xBB}}.Pack()))

	q4, err := queryparse.JSONStringToQuery(`{"four": 1000}`)
	assert.Nil(t, err)
	_, _, err = codec.ParseIndexQuery(q4)
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
