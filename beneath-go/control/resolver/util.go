package resolver

// DereferenceString does what it says it does
func DereferenceString(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}
