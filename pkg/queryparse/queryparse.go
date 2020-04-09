package queryparse

import (
	"fmt"
	"io"
	"strings"

	"gitlab.com/beneath-hq/beneath/pkg/jsonutil"
)

// ConditionOp is an enum representing possible conditions on fields
type ConditionOp int

// Query represents a parsed query
type Query map[string]*Condition

// IsEmpty returns true if the query filters no keys
func (q Query) IsEmpty() bool {
	return len(q) == 0
}

// Condition represents a condition on a single field in a query
type Condition struct {
	Op   ConditionOp
	Arg1 interface{}
	Arg2 interface{}
}

// ConditionOp enum definition
const (
	// ConditionOpNil is the nil condition (should never materialize)
	ConditionOpNil ConditionOp = iota
	// ConditionOpEq represents "_eq"
	ConditionOpEq
	// ConditionOpPrefix represents "_prefix"
	ConditionOpPrefix
	// ConditionOpGt represents "_gt"
	ConditionOpGt
	// ConditionOpGte represents "_gte"
	ConditionOpGte
	// ConditionOpLt represents "_lt"
	ConditionOpLt
	// ConditionOpLte represents "_lte"
	ConditionOpLte
	// ConditionOpGtLt represents "_gt" and "_lt"
	ConditionOpGtLt
	// ConditionOpGtLte represents "_gt" and "_lte"
	ConditionOpGtLte
	// ConditionOpGteLt represents "_gte" and "_lt"
	ConditionOpGteLt
	// ConditionOpGteLte represents "_gte" and "_lte"
	ConditionOpGteLte
)

// String converts the op to its string representation
func (o ConditionOp) String() string {
	switch o {
	case ConditionOpNil:
		return "<nil op>"
	case ConditionOpEq:
		return "_eq"
	case ConditionOpPrefix:
		return "_prefix"
	case ConditionOpGt:
		return "_gt"
	case ConditionOpGte:
		return "_gte"
	case ConditionOpLt:
		return "_lt"
	case ConditionOpLte:
		return "_lte"
	case ConditionOpGtLt:
		return "_gt,_lt"
	case ConditionOpGtLte:
		return "_gt,_lte"
	case ConditionOpGteLt:
		return "_gte,_lt"
	case ConditionOpGteLte:
		return "_gte,_lte"
	}
	return ""
}

// JSONStringToQuery is a wrapper for JSONToQuery
func JSONStringToQuery(json string) (Query, error) {
	if json != "" {
		return JSONReaderToQuery(strings.NewReader(json))
	}
	var query Query
	return query, nil
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

		// build condition
		if len(sq) == 1 { // single constraint
			if sq["_eq"] != nil {
				q[k] = &Condition{
					Op:   ConditionOpEq,
					Arg1: sq["_eq"],
				}
			} else if sq["_prefix"] != nil {
				q[k] = &Condition{
					Op:   ConditionOpPrefix,
					Arg1: sq["_prefix"],
				}
			} else if sq["_gt"] != nil {
				q[k] = &Condition{
					Op:   ConditionOpGt,
					Arg1: sq["_gt"],
				}
			} else if sq["_gte"] != nil {
				q[k] = &Condition{
					Op:   ConditionOpGte,
					Arg1: sq["_gte"],
				}
			} else if sq["_lt"] != nil {
				q[k] = &Condition{
					Op:   ConditionOpLt,
					Arg1: sq["_lt"],
				}
			} else if sq["_lte"] != nil {
				q[k] = &Condition{
					Op:   ConditionOpLte,
					Arg1: sq["_lte"],
				}
			}
		} else if len(sq) == 2 { // two constraints
			if sq["_gt"] != nil {
				if sq["_lt"] != nil {
					q[k] = &Condition{
						Op:   ConditionOpGtLt,
						Arg1: sq["_gt"],
						Arg2: sq["_lt"],
					}
				} else if sq["_lte"] != nil {
					q[k] = &Condition{
						Op:   ConditionOpGtLte,
						Arg1: sq["_gt"],
						Arg2: sq["_lte"],
					}
				}
			} else if sq["_gte"] != nil {
				if sq["_lt"] != nil {
					q[k] = &Condition{
						Op:   ConditionOpGteLt,
						Arg1: sq["_gte"],
						Arg2: sq["_lt"],
					}
				} else if sq["_lte"] != nil {
					q[k] = &Condition{
						Op:   ConditionOpGteLte,
						Arg1: sq["_gte"],
						Arg2: sq["_lte"],
					}
				}
			}
		}

		// error if nothing was assigned
		if q[k] == nil {
			return nil, fmt.Errorf("constraints for field '%s' not recognized", k)
		}
	}

	// done
	return q, nil
}
