package graphql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePrimitive(t *testing.T) {
	b, a, err := ParsePrimitive("Boolean")
	assert.Nil(t, err)
	assert.Equal(t, BooleanPrimitiveType, b)
	assert.Equal(t, 0, a)

	b, a, err = ParsePrimitive("Int31")
	assert.Equal(t, ErrNotPrimitive, err)

	b, a, err = ParsePrimitive("Int0")
	assert.Equal(t, ErrNotPrimitive, err)

	b, a, err = ParsePrimitive("Bytes128")
	assert.Nil(t, err)
	assert.Equal(t, BytesPrimitiveType, b)
	assert.Equal(t, 128, a)

	b, a, err = ParsePrimitive("Bytes")
	assert.Nil(t, err)
	assert.Equal(t, BytesPrimitiveType, b)
	assert.Equal(t, 0, a)

	isPrimitive := func(val string) bool {
		_, _, err := ParsePrimitive(val)
		return err != ErrNotPrimitive
	}

	assert.Equal(t, isPrimitive("Boolean"), true)
	assert.Equal(t, isPrimitive("Int31"), false)
	assert.Equal(t, isPrimitive("Int32"), true)
	assert.Equal(t, isPrimitive("Int"), true)
	assert.Equal(t, isPrimitive("Numeric"), true)
	assert.Equal(t, isPrimitive("String"), true)
	assert.Equal(t, isPrimitive("Timestamp"), true)
	assert.Equal(t, isPrimitive("Bytes0"), false)
	assert.Equal(t, isPrimitive("Bytes"), true)
	assert.Equal(t, isPrimitive("Bytes120"), true)
	assert.Equal(t, isPrimitive("Test"), false)
	assert.Equal(t, isPrimitive("[Bytes]"), false)
	assert.Equal(t, isPrimitive("[]Bytes"), false)
	assert.Equal(t, isPrimitive("Bytes[]"), false)
}
