package sdl

import "regexp"

var (
	primitiveRegex *regexp.Regexp
	indexableRegex *regexp.Regexp
)

func init() {
	primitiveRegex = regexp.MustCompile(makePrimitiveRegex())
	indexableRegex = regexp.MustCompile(makeIndexableRegex())
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

// returns true iff primitive
func isPrimitiveType(tr *TypeRef) bool {
	if tr.Array != nil {
		return false
	}
	return primitiveRegex.MatchString(tr.Type)
}

// returns true iff can be used in an index or as a key
func isIndexableType(tr *TypeRef) bool {
	if tr.Array != nil {
		return false
	}
	return indexableRegex.MatchString(tr.Type)
}
