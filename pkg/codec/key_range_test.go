package codec

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/beneath-core/pkg/codec/ext/tuple"
	"github.com/beneath-core/pkg/queryparse"
)

func TestKeyRange1(t *testing.T) {
	q, err := queryparse.JSONStringToQuery(`{ "a": "abc" }`)
	assert.Nil(t, err)

	index := testIndex{fields: []string{"a"}}
	c, err := New(`{"name":"test","type":"record","fields":[{"name": "a", "type": "string"}]}`, index, nil)
	assert.Nil(t, err)

	r, err := NewKeyRange(c, c.PrimaryIndex, q)
	assert.Nil(t, err)

	assert.True(t, r.IsPrefix())
	assert.True(t, r.Contains(tuple.Tuple{"abc"}.Pack()))
	assert.False(t, r.Contains(tuple.Tuple{"abcd"}.Pack()))
}

func TestKeyRange2(t *testing.T) {
	q, err := queryparse.JSONStringToQuery(`{ "a": { "_gt": 100 } }`)
	assert.Nil(t, err)

	index := testIndex{fields: []string{"a"}}
	c, err := New(`{"name":"test","type":"record","fields":[{"name": "a", "type": "long"}]}`, index, nil)
	assert.Nil(t, err)

	r, err := NewKeyRange(c, c.PrimaryIndex, q)
	assert.Nil(t, err)

	assert.False(t, r.IsPrefix())
	assert.False(t, r.Contains(tuple.Tuple{0}.Pack()))
	assert.False(t, r.Contains(tuple.Tuple{100}.Pack()))
	assert.True(t, r.Contains(tuple.Tuple{101}.Pack()))
	assert.True(t, r.Contains(tuple.Tuple{90876543}.Pack()))
}

func TestKeyRange3(t *testing.T) {
	q, err := queryparse.JSONStringToQuery(`{ "a": { "_gt": 100, "_lte": 200 } }`)
	assert.Nil(t, err)

	index := testIndex{fields: []string{"a"}}
	c, err := New(`{"name":"test","type":"record","fields":[{"name": "a", "type": "long"}]}`, index, nil)
	assert.Nil(t, err)

	r, err := NewKeyRange(c, c.PrimaryIndex, q)
	assert.Nil(t, err)

	assert.False(t, r.IsPrefix())
	assert.False(t, r.Contains(tuple.Tuple{100}.Pack()))
	assert.True(t, r.Contains(tuple.Tuple{150}.Pack()))
	assert.True(t, r.Contains(tuple.Tuple{200}.Pack()))
	assert.False(t, r.Contains(tuple.Tuple{201}.Pack()))
	assert.False(t, r.Contains(tuple.Tuple{90876543}.Pack()))
}

func TestKeyRange4(t *testing.T) {
	q, err := queryparse.JSONStringToQuery(`{ "a": 100, "b": { "_prefix": "ab" } }`)
	assert.Nil(t, err)

	index := testIndex{fields: []string{"a", "b"}}
	c, err := New(`{"name":"test","type":"record","fields":[{"name": "a", "type": "long"},{"name": "b", "type": "string"}]}`, index, nil)
	assert.Nil(t, err)

	r, err := NewKeyRange(c, c.PrimaryIndex, q)
	assert.Nil(t, err)

	assert.True(t, r.IsPrefix())
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

	index := testIndex{fields: []string{"a", "b"}}
	c, err := New(`{"name":"test","type":"record","fields":[{"name": "a", "type": "long"},{"name": "b", "type": "string"}]}`, index, nil)
	assert.Nil(t, err)

	_, err = NewKeyRange(c, c.PrimaryIndex, q)
	assert.NotNil(t, err)
	assert.Regexp(t, "cannot use '_prefix' on field 'a' because it only works on string and byte types", err.Error())
}

// func TestKeyRange6(t *testing.T) {
// 	where, err := queryparse.JSONStringToQuery(`{ "a": { "_eq": 100 } }`)
// 	assert.Nil(t, err)

// 	index := testIndex{fields: []string{"a", "b"}}
// 	c, err := New(`{"name":"test","type":"record","fields":[{"name": "a", "type": "long"},{"name": "b", "type": "string"}]}`, index, nil)
// 	assert.Nil(t, err)

// 	kr, err := NewKeyRange(c, where)
// 	assert.Nil(t, err)

// 	after, err := queryparse.JSONStringToQuery(`{ "a": 100 }`)
// 	assert.Nil(t, err)
// 	kr, err = kr.WithAfter(c, after)
// 	assert.NotNil(t, err)
// 	assert.Regexp(t, "after query must include exactly all keys fields and not more", err.Error())

// 	after, err = queryparse.JSONStringToQuery(`{ "a": 100, "b": {"_prefix": "bbb"} }`)
// 	assert.Nil(t, err)
// 	kr, err = kr.WithAfter(c, after)
// 	assert.NotNil(t, err)
// 	assert.Regexp(t, "after query cannot use '_prefix' constraint", err.Error())

// 	after, err = queryparse.JSONStringToQuery(`{ "a": 100, "b": "bbb" }`)
// 	assert.Nil(t, err)
// 	base1 := kr.Base
// 	kr, err = kr.WithAfter(c, after)
// 	assert.Nil(t, err)
// 	base2 := kr.Base
// 	assert.True(t, bytes.Compare(base1, base2) < 0)
// }

func TestKeyRange7(t *testing.T) {
	q, err := queryparse.JSONStringToQuery(`{ "a": { "_gt": 100, "_lte": 200 } }`)
	assert.Nil(t, err)

	index := testIndex{fields: []string{"a", "b"}}
	c, err := New(`{"name":"test","type":"record","fields":[{"name": "a", "type": "long"}, {"name": "b", "type": "long"}]}`, index, nil)
	assert.Nil(t, err)

	r, err := NewKeyRange(c, c.PrimaryIndex, q)
	assert.Nil(t, err)

	assert.False(t, r.IsPrefix())
	assert.False(t, r.Contains(tuple.Tuple{100, 10}.Pack()))
	assert.True(t, r.Contains(tuple.Tuple{150, 300}.Pack()))
	assert.True(t, r.Contains(tuple.Tuple{200}.Pack()))
	assert.True(t, r.Contains(tuple.Tuple{200, 300}.Pack()))
	assert.False(t, r.Contains(tuple.Tuple{201}.Pack()))
	assert.False(t, r.Contains(tuple.Tuple{90876543}.Pack()))
}
