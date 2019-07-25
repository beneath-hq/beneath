package codec

import (
	"testing"

	"github.com/beneath-core/beneath-go/core/codec/ext/tuple"

	"github.com/stretchr/testify/assert"

	"github.com/beneath-core/beneath-go/core/queryparse"
)

func TestKeyRange1(t *testing.T) {
	q, err := queryparse.JSONStringToQuery(`{ "a": "abc" }`)
	assert.Nil(t, err)

	c, err := NewKey([]string{"a"}, map[string]interface{}{
		"fields": []interface{}{
			map[string]interface{}{"name": "a", "type": "string"},
		},
	})
	assert.Nil(t, err)

	r, err := NewKeyRange(q, c)
	assert.Nil(t, err)

	assert.True(t, r.Contains(tuple.Tuple{"abc"}.Pack()))
	assert.False(t, r.Contains(tuple.Tuple{"abcd"}.Pack()))
}

func TestKeyRange2(t *testing.T) {
	q, err := queryparse.JSONStringToQuery(`{ "a": { "_gt": 100 } }`)
	assert.Nil(t, err)

	c, err := NewKey([]string{"a"}, map[string]interface{}{
		"fields": []interface{}{
			map[string]interface{}{"name": "a", "type": "long"},
		},
	})
	assert.Nil(t, err)

	r, err := NewKeyRange(q, c)
	assert.Nil(t, err)

	assert.False(t, r.Contains(tuple.Tuple{0}.Pack()))
	assert.False(t, r.Contains(tuple.Tuple{100}.Pack()))
	assert.True(t, r.Contains(tuple.Tuple{101}.Pack()))
	assert.True(t, r.Contains(tuple.Tuple{90876543}.Pack()))
}

func TestKeyRange3(t *testing.T) {
	q, err := queryparse.JSONStringToQuery(`{ "a": { "_gt": 100, "_lte": 200 } }`)
	assert.Nil(t, err)

	c, err := NewKey([]string{"a"}, map[string]interface{}{
		"fields": []interface{}{
			map[string]interface{}{"name": "a", "type": "long"},
		},
	})
	assert.Nil(t, err)

	r, err := NewKeyRange(q, c)
	assert.Nil(t, err)

	assert.False(t, r.Contains(tuple.Tuple{100}.Pack()))
	assert.True(t, r.Contains(tuple.Tuple{150}.Pack()))
	assert.True(t, r.Contains(tuple.Tuple{200}.Pack()))
	assert.False(t, r.Contains(tuple.Tuple{201}.Pack()))
	assert.False(t, r.Contains(tuple.Tuple{90876543}.Pack()))
}

func TestKeyRange4(t *testing.T) {
	q, err := queryparse.JSONStringToQuery(`{ "a": 100, "b": { "_prefix": "ab" } }`)
	assert.Nil(t, err)

	c, err := NewKey([]string{"a", "b"}, map[string]interface{}{
		"fields": []interface{}{
			map[string]interface{}{"name": "a", "type": "long"},
			map[string]interface{}{"name": "b", "type": "string"},
		},
	})
	assert.Nil(t, err)

	r, err := NewKeyRange(q, c)
	assert.Nil(t, err)

	assert.False(t, r.Contains(tuple.Tuple{100, "aab"}.Pack()))
	assert.True(t, r.Contains(tuple.Tuple{100, "ab"}.Pack()))
	assert.True(t, r.Contains(tuple.Tuple{100, "abz"}.Pack()))
	assert.False(t, r.Contains(tuple.Tuple{100, "ac"}.Pack()))
	assert.False(t, r.Contains(tuple.Tuple{99, "abz"}.Pack()))
	assert.False(t, r.Contains(tuple.Tuple{101, "abz"}.Pack()))
}

func TestKeyRange5(t *testing.T) {
	q, err := queryparse.JSONStringToQuery(`{ "a": { "_prefix": 100 } }`)
	assert.Nil(t, err)

	c, err := NewKey([]string{"a", "b"}, map[string]interface{}{
		"fields": []interface{}{
			map[string]interface{}{"name": "a", "type": "long"},
		},
	})
	assert.Nil(t, err)

	_, err = NewKeyRange(q, c)
	assert.NotNil(t, err)
	assert.Regexp(t, "cannot use '_prefix' on field 'a' because it only works on string and byte types", err.Error())
}
