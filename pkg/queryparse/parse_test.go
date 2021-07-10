package queryparse

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWhere(t *testing.T) {
	suite := []struct {
		in  string
		out Query
		err string
	}{
		{in: "foo = 100", out: Query{"foo": &Condition{Op: ConditionOpEq, Arg1: 100}}},
		{in: "foo == 100.20", out: Query{"foo": &Condition{Op: ConditionOpEq, Arg1: 100.2}}},
		{in: "foo is 100", out: Query{"foo": &Condition{Op: ConditionOpEq, Arg1: 100}}},
		{in: "foo starts with 'hello'", out: Query{"foo": &Condition{Op: ConditionOpPrefix, Arg1: "hello"}}},
		{in: "foo > 10", out: Query{"foo": &Condition{Op: ConditionOpGt, Arg1: 10}}},
		{in: "foo >= 10", out: Query{"foo": &Condition{Op: ConditionOpGte, Arg1: 10}}},
		{in: "foo < 10", out: Query{"foo": &Condition{Op: ConditionOpLt, Arg1: 10}}},
		{in: "foo <= 10", out: Query{"foo": &Condition{Op: ConditionOpLte, Arg1: 10}}},
		{in: "foo > 10, foo < 12", out: Query{"foo": &Condition{Op: ConditionOpGtLt, Arg1: 10, Arg2: 12}}},
		{in: "foo > 10, foo <= 12", out: Query{"foo": &Condition{Op: ConditionOpGtLte, Arg1: 10, Arg2: 12}}},
		{in: "foo >= 10, foo < 12", out: Query{"foo": &Condition{Op: ConditionOpGteLt, Arg1: 10, Arg2: 12}}},
		{in: "foo_XX >= 10, foo_XX <= 12", out: Query{"foo_XX": &Condition{Op: ConditionOpGteLte, Arg1: 10, Arg2: 12}}},
		{in: `foo > 10 , foo < 12 and bar starts with "foo"`, out: Query{"foo": &Condition{Op: ConditionOpGtLt, Arg1: 10, Arg2: 12}, "bar": &Condition{Op: ConditionOpPrefix, Arg1: "foo"}}},
		{in: "foo > 10, foo >= 10", err: "cannot combine operands .* for field 'foo'"},
		{in: "foo > 10, foo < 11, foo <= 15", err: "found more than two conditions on field 'foo'"},
		{in: `{"hello": {"_eq": "world"}}`, out: Query{"hello": &Condition{Op: ConditionOpEq, Arg1: "world"}}},
		{in: `{"foo": 100}`, out: Query{"foo": &Condition{Op: ConditionOpEq, Arg1: json.Number("100")}}},
		{in: `{"foo": {"_eq": "foo"}}`, out: Query{"foo": &Condition{Op: ConditionOpEq, Arg1: "foo"}}},
		{in: `{"foo": {"_prefix": "bar"}}`, out: Query{"foo": &Condition{Op: ConditionOpPrefix, Arg1: "bar"}}},
		{in: `{"foo": {"_gt": 200}, "bar": {"_gte": 200}}`, out: Query{"foo": &Condition{Op: ConditionOpGt, Arg1: json.Number("200")}, "bar": &Condition{Op: ConditionOpGte, Arg1: json.Number("200")}}},
		{in: `{"foo": {"_lt": 200}, "bar": {"_lte": 200}}`, out: Query{"foo": &Condition{Op: ConditionOpLt, Arg1: json.Number("200")}, "bar": &Condition{Op: ConditionOpLte, Arg1: json.Number("200")}}},
		{in: `{"foo": {"_gt": 200, "_lt": 300}}`, out: Query{"foo": &Condition{Op: ConditionOpGtLt, Arg1: json.Number("200"), Arg2: json.Number("300")}}},
		{in: `{"foo": {"_gt": 200, "_lte": 300}}`, out: Query{"foo": &Condition{Op: ConditionOpGtLte, Arg1: json.Number("200"), Arg2: json.Number("300")}}},
		{in: `{"foo": {"_gte": 200, "_lt": 300}}`, out: Query{"foo": &Condition{Op: ConditionOpGteLt, Arg1: json.Number("200"), Arg2: json.Number("300")}}},
		{in: `{"foo": {"_gte": 200, "_lte": 300}}`, out: Query{"foo": &Condition{Op: ConditionOpGteLte, Arg1: json.Number("200"), Arg2: json.Number("300")}}},
		{in: `{"hello":`, err: `query must be valid JSON`},
		{in: `{"hello": {"bad_constraint": "world"}}`, err: `invalid operand 'bad_constraint' for field 'hello'`},
		{in: `{"hello": {}}`, err: `empty conditions for field 'hello'`},
		{in: `{"hello": {"_gt": 10, "_gte": 11, "_lt": 15}}`, err: `too many conditions for field 'hello'`},
		{in: `{"hello": {"_lt": 10, "_lte": 11}}`, err: `cannot combine operands '_lte' and '_lt' for field 'hello'`},
	}

	for _, test := range suite {
		q, err := StringToQuery(test.in)
		if test.err == "" {
			assert.Nil(t, err)
			assert.Equal(t, len(test.out), len(q))
			for field, cond := range test.out {
				assert.NotNil(t, q[field])
				assert.Equal(t, cond.Op, q[field].Op)
				assert.Equal(t, cond.Arg1, q[field].Arg1)
				assert.Equal(t, cond.Arg2, q[field].Arg2)
			}
		} else {
			assert.NotNil(t, err)
			assert.Regexp(t, test.err, err.Error())
		}
	}
}
