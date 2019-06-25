package schema

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func parseAndValidate(schemaString string) error {
	var schema interface{}
	err := json.Unmarshal([]byte(schemaString), &schema)
	if err != nil {
		return err
	}
	return ValidateSchema(schema)
}

func Test1(t *testing.T) {
	err := parseAndValidate(`{
    "name": "test",
    "type": "record",
    "fields": [],
    "indexes": []
  }`)
	assert.NotNil(t, err)
	assert.Regexp(t, "record fields can't be empty", err.Error())
}

func Test2(t *testing.T) {
	err := parseAndValidate(`{
    "name": "test",
    "type": "record",
    "fields": [
      {"name": "one", "type": "long"}
    ],
    "indexes": [
    ]
  }`)
	assert.NotNil(t, err)
	assert.Regexp(t, "schema must specify at least one index", err.Error())
}

func Test3(t *testing.T) {
	err := parseAndValidate(`{
    "name": "test",
    "type": "record",
    "fields": [
      {"name": "one", "type": "long"}
    ],
    "indexes": [
      ["one"]
    ]
  }`)
	assert.Nil(t, err)
}

func Test4(t *testing.T) {
	err := parseAndValidate(`{
    "name": "test",
    "type": "record",
    "fields": [
      {"name": "one", "type": ["null", "long"]}
    ],
    "indexes": [
      ["one"]
    ]
  }`)
	assert.NotNil(t, err)
	assert.Regexp(t, "field <one> cannot be used in an index", err.Error())
}

func Test5(t *testing.T) {
	err := parseAndValidate(`{
    "name": "test",
    "type": "record",
    "fields": [
      {"name": "one", "type": {"type": "fixed", "size": 10}}
    ],
    "indexes": [
      ["one"]
    ]
  }`)
	assert.Nil(t, err)
}

func Test6(t *testing.T) {
	err := parseAndValidate(`{
    "name": "test",
    "type": "record",
    "fields": [
      {"name": "one", "type": "long"},
      {"name": "two", "type": {"type": "fixed", "size": 10, "logicalType": "decimal"}}
    ],
    "indexes": [
      ["one"]
    ]
  }`)
	assert.NotNil(t, err)
	assert.Regexp(t, "logicalType not supported: fixed.decimal", err.Error())
}

func Test7(t *testing.T) {
	err := parseAndValidate(`{
    "name": "test",
    "type": "record",
    "fields": [
      {"name": "one", "type": "long"},
      {"name": "two", "type": "bytes", "logicalType": "decimal"}
    ],
    "indexes": [
      ["one"]
    ]
  }`)
	assert.Nil(t, err)
}

func Test8(t *testing.T) {
	err := parseAndValidate(`{
    "name": "test",
    "type": "record",
    "fields": [
      {"name": "one", "type": "string"},
      {"name": "two", "type": ["null", "boolean"]}
    ],
    "indexes": [
      ["one"]
    ]
  }`)
	assert.Nil(t, err)
}

func Test9(t *testing.T) {
	err := parseAndValidate(`{
    "name": "test",
    "type": "record",
    "fields": [
      {"name": "one", "type": "string"},
      {"name": "two", "type": ["null", "bytes", "int"]}
    ],
    "indexes": [
      ["one"]
    ]
  }`)
	assert.NotNil(t, err)
	assert.Regexp(t, "unions can have only two types and the first must be 'null'", err.Error())
}

func Test10(t *testing.T) {
	err := parseAndValidate(`{
    "name": "test",
    "type": "record",
    "fields": [
      {"name": "one", "type": "string"},
      {"name": "two", "type": ["null", "test"]}
    ],
    "indexes": [
      ["one"]
    ]
  }`)
	assert.NotNil(t, err)
	assert.Regexp(t, "type not supported: test", err.Error())
}

func Test11(t *testing.T) {
	err := parseAndValidate(`{
    "name": "test",
    "type": "record",
    "fields": [
      {"name": "one", "type": "string"},
      {"name": "two", "type": "array", "items": ["null", {
        "name": "test_2",
        "type": "record",
        "fields": [
          {"name": "one", "type": "double"}
        ]
      }]}
    ],
    "indexes": [
      ["one"]
    ]
  }`)
	assert.Nil(t, err)
}

func Test12(t *testing.T) {
	err := parseAndValidate(`{
    "name": "test",
    "type": "record",
    "fields": [
      {"name": "one", "type": "string"},
      {"name": "two", "type": "array", "items": ["null", {
        "name": "test_2",
        "type": "record",
        "fields": [
          {"name": "one", "type": "double"}
        ]
      }]}
    ],
    "indexes": [
      ["one", "two"]
    ]
  }`)
	assert.NotNil(t, err)
	assert.Regexp(t, "field <two> cannot be used in an index", err.Error())
}
