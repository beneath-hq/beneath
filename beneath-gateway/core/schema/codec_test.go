package schema

import (
	"encoding/json"
	"testing"

	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"
)

func TestSimple(t *testing.T) {
	codec, err := NewCodec(`{
		"name": "test",
		"type": "record",
		"fields": [
			{"name": "key_one", "type": "string"},
			{"name": "val_one", "type": "string"},
			{"name": "val_two", "type": "long"}
		]
	}`, [][]string{[]string{"key_one"}})
	assert.Nil(t, err)

	valueJSON := []byte(`{
		"key_one": "testing three two one",
		"val_one": "0xItLooksLikeAHexButYouGotTricked",
		"val_two": "-9007199254741992"
	}`)

	var jsonNative, jsonNativeCopy interface{}
	err = json.Unmarshal([]byte(valueJSON), &jsonNative)
	assert.Nil(t, err)
	err = json.Unmarshal([]byte(valueJSON), &jsonNativeCopy)
	assert.Nil(t, err)

	binary, err := codec.Marshal(jsonNative)
	assert.Nil(t, err)

	native, err := codec.Unmarshal(binary)

	if diff := deep.Equal(jsonNativeCopy, native); diff != nil {
		t.Error(diff)
	}
}

func TestComplex(t *testing.T) {
	codec, err := NewCodec(`{
		"name": "test",
		"type": "record",
		"fields": [
			{"name": "key_one", "type": "string"},
			{"name": "key_two", "type": "fixed", "size": 20},
			{"name": "key_three", "type": "long"},
			{"name": "val_one", "type": "string"},
			{"name": "val_two", "type": "bytes", "logicalType": "decimal", "precision": 78},
			{"name": "val_three", "type": "long", "logicalType": "timestamp-millis"},
			{"name": "val_four", "type": "array", "items": {
				"name": "val_four",
				"type": "record",
				"fields": [
					{"name": "val_four_one", "type": "boolean"},
					{"name": "val_four_two", "type": "double"}
				]
			}}
		]
	}`, [][]string{[]string{"key_one", "key_two", "key_three"}})
	assert.Nil(t, err)

	valueJSON := `{
		"key_one": "testing one two three",
		"key_two": "0xaabbccddeeaabbccddeeaabbccddeeaabbccddee",
		"key_three": 1234567890,
		"val_one": "There was a bell beside the gate, and Dorothy pushed the button and heard a silvery tinkle sound within. Then the big gate swung slowly open, and they all passed through and found themselves in a high arched room, the walls of which glistened with countless emeralds. Before them stood a little man about the same size as the Munchkins. He was clothed all in green, from his head to his feet, and even his skin was of a greenish tint. At his side was a large green box.",
		"val_two": "-77224998599806363752588771300231266558642741460645341489178111450841839741627",
		"val_three": 1560949036000,
		"val_four": [
			{ "val_four_one": true, "val_four_two": 3.14159265358 },
			{ "val_four_one": false, "val_four_two": 2.718281828 }
		]
	}`

	var jsonNative, jsonNativeCopy interface{}
	err = json.Unmarshal([]byte(valueJSON), &jsonNative)
	assert.Nil(t, err)
	err = json.Unmarshal([]byte(valueJSON), &jsonNativeCopy)
	assert.Nil(t, err)

	binary, err := codec.Marshal(jsonNative)
	assert.Nil(t, err)

	native, err := codec.Unmarshal(binary)
	assert.Nil(t, err)

	if diff := deep.Equal(jsonNativeCopy, native); diff != nil {
		t.Error(diff)
	}
}
