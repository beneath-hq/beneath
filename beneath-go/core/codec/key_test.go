package codec

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func hexToBytes(num string) []byte {
	bytes, err := hex.DecodeString(strings.Replace(num, "0x", "", 1))
	if err != nil {
		panic(err.Error())
	}
	return bytes
}

func TestKeySimple(t *testing.T) {
	keyFields := []string{"k1", "k2", "k3", "k4"}
	schemaString := `{
		"name": "test",
		"type": "record",
		"fields": [
			{"name": "k1", "type": "fixed", "size": 20},
			{"name": "k2", "type": "long"},
			{"name": "k3", "type": "string"},
			{"name": "k4", "type": "long", "logicalType": "timestamp-millis"}
		]
	}`

	avroCodec, err := NewAvro(schemaString)
	assert.Nil(t, err)

	keyCodec, err := NewKey(keyFields, avroCodec.GetSchema())
	assert.Nil(t, err)

	v1, err := keyCodec.Marshal(map[string]interface{}{
		"k1": hexToBytes("0x0000000000000000000000000000000000000000"),
		"k2": 10000000000000,
		"k3": "abc",
		"k4": time.Date(1995, time.February, 1, 0, 0, 0, 0, time.UTC),
	})
	assert.Nil(t, err)
	assert.NotNil(t, v1)

	v2, err := keyCodec.Marshal(map[string]interface{}{
		"k1": hexToBytes("0x0000000000000000000000000000000000000000"),
		"k2": 10000000000000,
		"k3": "abc",
		"k4": time.Date(1995, time.February, 2, 0, 0, 0, 0, time.UTC),
	})
	assert.Nil(t, err)
	assert.NotNil(t, v2)

	v3, err := keyCodec.Marshal(map[string]interface{}{
		"k1": hexToBytes("0x0000000000000000000000000000000000000000"),
		"k2": 10000000000000,
		"k3": "abcd",
		"k4": time.Date(1995, time.February, 1, 0, 0, 0, 0, time.UTC),
	})
	assert.Nil(t, err)
	assert.NotNil(t, v3)

	v4, err := keyCodec.Marshal(map[string]interface{}{
		"k1": hexToBytes("0x0000000000000000000000000000000000000000"),
		"k2": 90000000000000,
		"k3": "abc",
		"k4": time.Date(1995, time.February, 1, 0, 0, 0, 0, time.UTC),
	})
	assert.Nil(t, err)
	assert.NotNil(t, v4)

	v5, err := keyCodec.Marshal(map[string]interface{}{
		"k1": hexToBytes("0xFF00000000000000000000000000000000000000"),
		"k2": 10000000000000,
		"k3": "abc",
		"k4": time.Date(1995, time.February, 1, 0, 0, 0, 0, time.UTC),
	})
	assert.Nil(t, err)
	assert.NotNil(t, v5)

	_, err = keyCodec.Marshal(map[string]interface{}{
		"k1": hexToBytes("0xFF00000000000000000000000000000000000000"),
		"k2": 10000000000000,
		"k4": time.Date(1995, time.February, 1, 0, 0, 0, 0, time.UTC),
	})
	assert.NotNil(t, err)
	assert.Equal(t, "Value for key field 'k3' is nil", err.Error())

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
