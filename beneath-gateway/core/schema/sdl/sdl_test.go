package sdl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSDL1(t *testing.T) {
	c := NewCompiler(`
		type Test1 @stream(name: "eth-blocks", key: ["blockNumber", "timestamp"]) {
			blockNumber: Int!
			blockHash: String
			timestamp: Int!
		}
	`)
	err := c.Compile()
	assert.Nil(t, err)
}
