package sdl

import (
	"testing"

	"github.com/linkedin/goavro/v2"
	"github.com/stretchr/testify/assert"
)

func TestAvro1(t *testing.T) {
	c := NewCompiler(`
		type TestA @stream(name: "test-example-1", key: ["aAaa", "aBbb"]) {
			aAaa: String!
			aBbb: Timestamp!
			aCcc: [TestB]
		}
		type TestB {
			bAaa: Int
			bBbb: Bytes
			bCcc: TestC!
		}
		enum TestC {
			Aaa
			Bbb
			Ccc
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
}
