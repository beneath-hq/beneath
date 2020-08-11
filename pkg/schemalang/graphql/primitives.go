package graphql

import (
	"fmt"
	"regexp"
	"strconv"
)

// PrimitiveType represents a GraphQL primitive type
type PrimitiveType string

// Constants for PrimitiveType
const (
	BooleanPrimitiveType   PrimitiveType = "Boolean"
	BytesPrimitiveType     PrimitiveType = "Bytes"
	FloatPrimitiveType     PrimitiveType = "Float"
	IntPrimitiveType       PrimitiveType = "Int"
	NumericPrimitiveType   PrimitiveType = "Numeric"
	StringPrimitiveType    PrimitiveType = "String"
	TimestampPrimitiveType PrimitiveType = "Timestamp"
	UUIDPrimitiveType      PrimitiveType = "UUID"
)

// Primitives is a slice of the constants for PrimitiveType
var Primitives = []PrimitiveType{
	BooleanPrimitiveType,
	BytesPrimitiveType,
	FloatPrimitiveType,
	IntPrimitiveType,
	NumericPrimitiveType,
	StringPrimitiveType,
	TimestampPrimitiveType,
	UUIDPrimitiveType,
}

// ErrNotPrimitive is returned from ParsePrimitive for values that are not recognized primitives
var ErrNotPrimitive = fmt.Errorf("type name is not a GraphQL primitive type")

// used to split suffix numbers from a primitive (e.g. Bytes32 -> (Bytes, 32))
var splitPrimitiveNumbersRegex = regexp.MustCompile("^([a-zA-Z]+)([1-9][0-9]*)?$")

// ParsePrimitive parses numbers from a primitive, e.g. 'Bytes32' -> ('Bytes', 32)
func ParsePrimitive(val string) (base PrimitiveType, arg int, err error) {
	matches := splitPrimitiveNumbersRegex.FindStringSubmatch(val)
	if len(matches) != 3 {
		return "", 0, ErrNotPrimitive
	}

	base = PrimitiveType(matches[1])
	arg, _ = strconv.Atoi(matches[2])

	var valid bool
	switch base {
	case BooleanPrimitiveType:
		valid = arg == 0
	case BytesPrimitiveType:
		valid = true
	case FloatPrimitiveType:
		valid = arg == 0 || arg == 32 || arg == 64
	case IntPrimitiveType:
		valid = arg == 0 || arg == 32 || arg == 64
	case NumericPrimitiveType:
		valid = arg == 0
	case StringPrimitiveType:
		valid = arg == 0
	case TimestampPrimitiveType:
		valid = arg == 0
	case UUIDPrimitiveType:
		valid = arg == 0
	default:
		valid = false
	}

	if !valid {
		return "", 0, ErrNotPrimitive
	}

	return base, arg, nil
}
