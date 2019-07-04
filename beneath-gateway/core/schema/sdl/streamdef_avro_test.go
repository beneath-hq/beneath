package sdl

import (
	"testing"

	"github.com/linkedin/goavro/v2"
	"github.com/stretchr/testify/assert"
)

// Note: use fixed20 twice and see if name causes error

func TestAvro1(t *testing.T) {
	c := NewCompiler(`
		type TestA @stream(name: "test", key: ["a", "b"]) {
			a: String!
			b: Timestamp!
			c: [TestB!]
			d: TestC!
			e: Bytes20
		}
		type TestB {
			a: Int
			b: Bytes
			c: TestC!
			d: Bytes20
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
	// log.Printf(avro)
	assert.Nil(t, err)
	assert.NotNil(t, avro)

	codec, err := goavro.NewCodec(avro)
	assert.Nil(t, err)
	assert.NotNil(t, codec)
}
