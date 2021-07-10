package queryparse

import (
	"fmt"
	"io"
	"strings"

	"github.com/beneath-hq/beneath/pkg/jsonutil"
)

// Maps JSON keys to ops
var keyToOp = map[string]ConditionOp{
	"=":       ConditionOpEq,
	"==":      ConditionOpEq,
	"eq":      ConditionOpEq,
	"_eq":     ConditionOpEq,
	"prefix":  ConditionOpPrefix,
	"_prefix": ConditionOpPrefix,
	">":       ConditionOpGt,
	"gt":      ConditionOpGt,
	"_gt":     ConditionOpGt,
	">=":      ConditionOpGte,
	"gte":     ConditionOpGte,
	"_gte":    ConditionOpGte,
	"<":       ConditionOpLt,
	"lt":      ConditionOpLt,
	"_lt":     ConditionOpLt,
	"<=":      ConditionOpLte,
	"lte":     ConditionOpLte,
	"_lte":    ConditionOpLte,
}

// JSONStringToQuery is a wrapper for JSONToQuery
func JSONStringToQuery(json string) (Query, error) {
	if strings.TrimSpace(json) == "" {
		var query Query
		return query, nil
	}
	return JSONReaderToQuery(strings.NewReader(json))
}

// JSONReaderToQuery is a wrapper for JSONToQuery
func JSONReaderToQuery(reader io.Reader) (Query, error) {
	var json map[string]interface{}
	err := jsonutil.Unmarshal(reader, &json)
	if err != nil {
		return nil, fmt.Errorf("query must be valid JSON")
	}
	return JSONToQuery(json)
}

// JSONToQuery creates a new query from parsed json
func JSONToQuery(json map[string]interface{}) (Query, error) {
	// iterate through json and build query
	q := make(Query, len(json))
	for k, sqT := range json {
		// cast subquery to map
		sq, ok := sqT.(map[string]interface{})
		if !ok {
			// if it's not a map, we'll treat it as a direct value comparison
			sq = map[string]interface{}{
				"_eq": sqT,
			}
		}

		// there can be only one or two operands for a field
		if len(sq) == 0 {
			return nil, fmt.Errorf("empty conditions for field '%s'", k)
		} else if len(sq) > 2 {
			return nil, fmt.Errorf("too many conditions for field '%s'", k)
		}

		// build condition
		cond := &Condition{}
		for literal, arg := range sq {
			// map to operand
			op := keyToOp[literal]
			if op == ConditionOpNil {
				return nil, fmt.Errorf("invalid operand '%s' for field '%s'", literal, k)
			}

			// if it's the first op, just add to cond
			if cond.Op == ConditionOpNil {
				// first op
				cond.Op = op
				cond.Arg1 = arg
				continue
			}

			// this is a second operand, we need to merge it into cond
			cond.Arg2 = arg
			if cond.Op == ConditionOpLt || cond.Op == ConditionOpLte {
				// flip so that cond.Op and cond.Arg1 are not lt/lte
				tmpOp := cond.Op
				cond.Op = op
				op = tmpOp
				tmpArg := cond.Arg1
				cond.Arg1 = cond.Arg2
				cond.Arg2 = tmpArg
			}
			// merge ops, knowing cond.Op cannot be lt/lte
			if cond.Op == ConditionOpGt && op == ConditionOpLt {
				cond.Op = ConditionOpGtLt
			} else if cond.Op == ConditionOpGt && op == ConditionOpLte {
				cond.Op = ConditionOpGtLte
			} else if cond.Op == ConditionOpGte && op == ConditionOpLt {
				cond.Op = ConditionOpGteLt
			} else if cond.Op == ConditionOpGte && op == ConditionOpLte {
				cond.Op = ConditionOpGteLte
			} else {
				return nil, fmt.Errorf("cannot combine operands '%s' and '%s' for field '%s'", cond.Op, op, k)
			}
		}
		q[k] = cond
	}

	// done
	return q, nil
}
