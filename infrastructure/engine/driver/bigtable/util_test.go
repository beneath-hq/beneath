package bigtable

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntToBytes(t *testing.T) {
	xs := []int64{0, 1, 2, -1, -2, 314, -314, 0xFFFFFFFF, -0xFFFFFFFF, 0x7FFFFFFFFFFFFFFF, -0x8000000000000000}
	for _, x := range xs {
		b := intToBytes(x)
		y := bytesToInt(b)
		assert.Equal(t, x, y)
		assert.Equal(t, 8, len(b))
	}
}

func TestSplitCommonPrefix(t *testing.T) {
	a := []byte{0x11, 0x22, 0x33}
	b := []byte{0x11, 0x22, 0x44}
	p, aa, bb := splitCommonPrefix(a, b)
	assert.True(t, bytes.Equal(p, []byte{0x11, 0x22}))
	assert.True(t, bytes.Equal(aa, []byte{0x33}))
	assert.True(t, bytes.Equal(bb, []byte{0x44}))

	a = []byte{0x11, 0x22, 0x33}
	b = []byte{0x99, 0x88, 0x77}
	p, aa, bb = splitCommonPrefix(a, b)
	assert.Equal(t, 0, len(p))
	assert.True(t, bytes.Equal(aa, a))
	assert.True(t, bytes.Equal(bb, b))

	a = []byte{0x11, 0x22}
	b = []byte{0x11, 0x22, 0x33, 0x44}
	p, aa, bb = splitCommonPrefix(a, b)
	assert.True(t, bytes.Equal(p, []byte{0x11, 0x22}))
	assert.Equal(t, 0, len(aa))
	assert.True(t, bytes.Equal(bb, []byte{0x33, 0x44}))
}
