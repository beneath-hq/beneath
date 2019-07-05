package schema

import (
	"testing"

	"cloud.google.com/go/bigquery"
	"github.com/stretchr/testify/assert"
)

func TestBigQuery1(t *testing.T) {
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

	json, err := s.BuildBigQuerySchema()
	assert.Nil(t, err)
	assert.NotNil(t, json)

	schema, err := bigquery.SchemaFromJSON([]byte(json))
	assert.Nil(t, err)
	assert.NotNil(t, schema)
}
