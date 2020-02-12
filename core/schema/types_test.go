package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsPrimitiveTypeName(t *testing.T) {
	assert.Equal(t, isPrimitiveTypeName("Boolean"), true)
	assert.Equal(t, isPrimitiveTypeName("Int31"), false)
	assert.Equal(t, isPrimitiveTypeName("Int32"), true)
	assert.Equal(t, isPrimitiveTypeName("Int"), true)
	assert.Equal(t, isPrimitiveTypeName("Numeric"), true)
	assert.Equal(t, isPrimitiveTypeName("String"), true)
	assert.Equal(t, isPrimitiveTypeName("Timestamp"), true)
	assert.Equal(t, isPrimitiveTypeName("Bytes0"), false)
	assert.Equal(t, isPrimitiveTypeName("Bytes"), true)
	assert.Equal(t, isPrimitiveTypeName("Bytes120"), true)
	assert.Equal(t, isPrimitiveTypeName("Test"), false)
	assert.Equal(t, isPrimitiveTypeName("[Bytes]"), false)
	assert.Equal(t, isPrimitiveTypeName("[]Bytes"), false)
	assert.Equal(t, isPrimitiveTypeName("Bytes[]"), false)
}

func TestIsIndexableTypeName(t *testing.T) {
	// isIndexableTypeName
	assert.Equal(t, isIndexableTypeName("Boolean"), false)
	assert.Equal(t, isIndexableTypeName("Int31"), false)
	assert.Equal(t, isIndexableTypeName("Int32"), true)
	assert.Equal(t, isIndexableTypeName("Int"), true)
	assert.Equal(t, isIndexableTypeName("Numeric"), false)
	assert.Equal(t, isIndexableTypeName("String"), true)
	assert.Equal(t, isIndexableTypeName("Timestamp"), true)
	assert.Equal(t, isIndexableTypeName("Bytes0"), false)
	assert.Equal(t, isIndexableTypeName("Bytes"), true)
	assert.Equal(t, isIndexableTypeName("Bytes120"), true)
	assert.Equal(t, isIndexableTypeName("Test"), false)
	assert.Equal(t, isIndexableTypeName("[Bytes]"), false)
	assert.Equal(t, isIndexableTypeName("[]Bytes"), false)
	assert.Equal(t, isIndexableTypeName("Bytes[]"), false)
}

func TestSplitPrimitiveName(t *testing.T) {
	b, a := splitPrimitiveName("Boolean")
	assert.Equal(t, "Boolean", b)
	assert.Equal(t, 0, a)

	b, a = splitPrimitiveName("Int31")
	assert.Equal(t, "Int", b)
	assert.Equal(t, 31, a)

	b, a = splitPrimitiveName("Int0")
	assert.Equal(t, "", b)
	assert.Equal(t, 0, a)

	b, a = splitPrimitiveName("Bytes128")
	assert.Equal(t, "Bytes", b)
	assert.Equal(t, 128, a)

	b, a = splitPrimitiveName("Bytes")
	assert.Equal(t, "Bytes", b)
	assert.Equal(t, 0, a)
}
