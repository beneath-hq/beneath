package queryparse

import (
	"fmt"
	"strings"

	"github.com/beneath-hq/beneath/pkg/queryparse/whereparser"
)

// Maps AST ops to ConditionOp
var astToOp = map[string]ConditionOp{
	"=":           ConditionOpEq,
	"==":          ConditionOpEq,
	"IS":          ConditionOpEq,
	"STARTS WITH": ConditionOpPrefix,
	">":           ConditionOpGt,
	">=":          ConditionOpGte,
	"<":           ConditionOpLt,
	"<=":          ConditionOpLte,
}

// JSONStringToQuery is a wrapper for JSONToQuery
func WhereStringToQuery(where string) (Query, error) {
	// return if there is no query
	if strings.TrimSpace(where) == "" {
		var query Query
		return query, nil
	}

	// parse into AST
	ast, err := whereparser.Parse(where)
	if err != nil {
		return nil, err
	}

	// build query
	q := make(Query, len(ast.Conditions))
	for _, node := range ast.Conditions {
		// get op
		op := astToOp[node.Op]
		if op == ConditionOpNil {
			return nil, fmt.Errorf("invalid operand '%s' for field '%s'", node.Op, node.Field)
		}

		// get the arg
		var arg interface{}
		if node.Value.String != nil {
			arg = *node.Value.String
		} else if node.Value.Number != nil {
			arg = *node.Value.Number
		} else if node.Value.Boolean != nil {
			arg = bool(*node.Value.Boolean)
		} else if node.Value.Null {
			arg = nil
		}

		// create the cond
		cond := q[node.Field]
		if cond == nil {
			// field not seen before
			q[node.Field] = &Condition{
				Op:   op,
				Arg1: arg,
			}
		} else {
			// this is a second operand on the field, we need to merge it into cond

			// first check it's not a third op
			if cond.Op == ConditionOpGtLt || cond.Op == ConditionOpGtLte || cond.Op == ConditionOpGteLt || cond.Op == ConditionOpGteLte {
				return nil, fmt.Errorf("found more than two conditions on field '%s'", node.Field)
			}

			// merge
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
				return nil, fmt.Errorf("cannot combine operands '%s' and '%s' for field '%s'", cond.Op, op, node.Field)
			}
		}
	}

	return q, nil
}
