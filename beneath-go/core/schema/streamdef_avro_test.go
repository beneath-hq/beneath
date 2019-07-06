package schema

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/linkedin/goavro/v2"
	"github.com/stretchr/testify/assert"
)

func TestAvro1(t *testing.T) {
	c := NewCompiler(`
		" Docs, docs, docs, docs! "
		type TestA @stream(name: "test", key: ["a", "b"]) {
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

	err := c.Compile()
	assert.Nil(t, err)

	s := c.GetStream()
	assert.NotNil(t, s)

	avro, err := s.BuildAvroSchema(true)
	assert.Nil(t, err)
	assert.NotNil(t, avro)

	codec, err := goavro.NewCodec(avro)
	assert.Nil(t, err)
	assert.NotNil(t, codec)

	json1 := `{
		"a": "Example example",
		"b": 1562270699000,
		"c": { "Bytes20": "aaaaaaaaaaaaaaaaaaaa" },
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
