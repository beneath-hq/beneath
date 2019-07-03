package sdl

// returns true iff primitive
func isPrimitiveType(tr *TypeRef) bool {
	if tr.Array != nil {
		return false
	}

	// TODO

	return (tr.Type == "Boolean" ||
		tr.Type == "Int" ||
		tr.Type == "Int32" ||
		tr.Type == "Int64" ||
		tr.Type == "Float" ||
		tr.Type == "Float32" ||
		tr.Type == "Float64" ||
		tr.Type == "Numeric" ||
		tr.Type == "Timestamp" ||
		tr.Type == "Bytes" ||
		tr.Type == "Bytes20" ||
		tr.Type == "String")
}

// returns true iff can be used in an index or as a key
func isIndexableType(tr *TypeRef) bool {
	return isPrimitiveType(tr) && (tr.Type == "Int" ||
		tr.Type == "Int32" ||
		tr.Type == "Int64" ||
		tr.Type == "Timestamp" ||
		tr.Type == "Bytes" ||
		tr.Type == "String")
}
