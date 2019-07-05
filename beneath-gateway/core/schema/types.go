package schema

import (
	"regexp"
	"strconv"
)

var (
	primitiveRegex *regexp.Regexp
	indexableRegex *regexp.Regexp
	splitRegex     *regexp.Regexp
)

func init() {
	primitiveRegex = regexp.MustCompile(makePrimitiveRegex())
	indexableRegex = regexp.MustCompile(makeIndexableRegex())
	splitRegex = regexp.MustCompile(makeSplitRegex())
}

func makePrimitiveRegex() string {
	return ("^(()" +
		"|(Boolean)" +
		"|(Bytes)" +
		"|(Bytes[1-9][0-9]*)" +
		"|(Float)" +
		"|(Float32)" +
		"|(Float64)" +
		"|(Int)" +
		"|(Int32)" +
		"|(Int64)" +
		"|(Numeric)" +
		"|(String)" +
		"|(Timestamp)" +
		")$")
}

func makeIndexableRegex() string {
	return ("^(()" +
		"|(Bytes)" +
		"|(Bytes[1-9][0-9]*)" +
		"|(Int)" +
		"|(Int32)" +
		"|(Int64)" +
		"|(String)" +
		"|(Timestamp)" +
		")$")
}

func makeSplitRegex() string {
	return "^([a-zA-Z]+)([1-9][0-9]*)?$"
}

// returns true iff primitive
func isPrimitiveType(tr *TypeRef) bool {
	if tr.Array != nil {
		return false
	}
	return isPrimitiveTypeName(tr.Type)
}

// same as isPrimitiveType, but for type string
func isPrimitiveTypeName(n string) bool {
	return primitiveRegex.MatchString(n)
}

// returns true iff can be used in an index or as a key
func isIndexableType(tr *TypeRef) bool {
	if tr.Array != nil {
		return false
	}
	return isIndexableTypeName(tr.Type)
}

// same as isIndexableType, but for type string
func isIndexableTypeName(n string) bool {
	return indexableRegex.MatchString(n)
}

func splitPrimitiveName(name string) (base string, arg int) {
	matches := splitRegex.FindStringSubmatch(name)
	if len(matches) == 3 {
		base = matches[1]
		arg, _ = strconv.Atoi(matches[2])
	}
	return base, arg
}
