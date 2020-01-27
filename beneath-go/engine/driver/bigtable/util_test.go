package bigtable

import (
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
