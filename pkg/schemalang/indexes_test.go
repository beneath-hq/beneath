package schemalang

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Note: there's a fair bunch of tests that touch Check in the transpiler test cases.

func TestIndexesCanonicalJSON(t *testing.T) {
	first := Indexes{
		Index{Fields: []string{"aaa", "bbb"}, Key: false},
		Index{Fields: []string{"aaa", "ccc"}, Key: true},
		Index{Fields: []string{"ddd"}, Key: false, Normalize: true},
	}

	second := Indexes{
		Index{Fields: []string{"aaa", "ccc"}, Key: true},
		Index{Fields: []string{"ddd"}, Key: false, Normalize: true},
		Index{Fields: []string{"aaa", "bbb"}, Key: false},
	}

	expected := `[{"fields":["aaa","ccc"],"key":true},{"fields":["aaa","bbb"]},{"fields":["ddd"],"normalize":true}]`
	assert.Equal(t, expected, first.CanonicalJSON())
	assert.Equal(t, expected, second.CanonicalJSON())
}
