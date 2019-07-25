package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuery(t *testing.T) {
	q, err := JSONToQuery(map[string]interface{}{
		"eq1":     100,
		"eq2":     map[string]interface{}{"_eq": "foo"},
		"prefix":  map[string]interface{}{"_prefix": "bar"},
		"gt":      map[string]interface{}{"_gt": 200},
		"gte":     map[string]interface{}{"_gte": 200},
		"lt":      map[string]interface{}{"_lt": 200},
		"lte":     map[string]interface{}{"_lte": 200},
		"gt,lt":   map[string]interface{}{"_gt": 200, "_lt": 300},
		"gt,lte":  map[string]interface{}{"_gt": 200, "_lte": 300},
		"gte,lt":  map[string]interface{}{"_gte": 200, "_lt": 300},
		"gte,lte": map[string]interface{}{"_gte": 200, "_lte": 300},
	})
	assert.Nil(t, err)

	assert.Equal(t, ConditionOpEq, q["eq1"].Op)
	assert.Equal(t, 100, q["eq1"].Arg1)
	assert.Equal(t, nil, q["eq1"].Arg2)

	assert.Equal(t, ConditionOpEq, q["eq2"].Op)
	assert.Equal(t, "foo", q["eq2"].Arg1)
	assert.Equal(t, nil, q["eq2"].Arg2)

	assert.Equal(t, ConditionOpPrefix, q["prefix"].Op)
	assert.Equal(t, "bar", q["prefix"].Arg1)
	assert.Equal(t, nil, q["prefix"].Arg2)

	assert.Equal(t, ConditionOpGt, q["gt"].Op)
	assert.Equal(t, 200, q["gt"].Arg1)
	assert.Equal(t, nil, q["gt"].Arg2)

	assert.Equal(t, ConditionOpGte, q["gte"].Op)
	assert.Equal(t, 200, q["gte"].Arg1)
	assert.Equal(t, nil, q["gte"].Arg2)

	assert.Equal(t, ConditionOpLt, q["lt"].Op)
	assert.Equal(t, 200, q["lt"].Arg1)
	assert.Equal(t, nil, q["lt"].Arg2)

	assert.Equal(t, ConditionOpLte, q["lte"].Op)
	assert.Equal(t, 200, q["lte"].Arg1)
	assert.Equal(t, nil, q["lte"].Arg2)

	assert.Equal(t, ConditionOpGtLt, q["gt,lt"].Op)
	assert.Equal(t, 200, q["gt,lt"].Arg1)
	assert.Equal(t, 300, q["gt,lt"].Arg2)

	assert.Equal(t, ConditionOpGtLte, q["gt,lte"].Op)
	assert.Equal(t, 200, q["gt,lte"].Arg1)
	assert.Equal(t, 300, q["gt,lte"].Arg2)

	assert.Equal(t, ConditionOpGteLt, q["gte,lt"].Op)
	assert.Equal(t, 200, q["gte,lt"].Arg1)
	assert.Equal(t, 300, q["gte,lt"].Arg2)

	assert.Equal(t, ConditionOpGteLte, q["gte,lte"].Op)
	assert.Equal(t, 200, q["gte,lte"].Arg1)
	assert.Equal(t, 300, q["gte,lte"].Arg2)
}
