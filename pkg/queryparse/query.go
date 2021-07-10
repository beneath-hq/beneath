package queryparse

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

// ConditionOp is an enum representing possible conditions on fields
type ConditionOp int

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
