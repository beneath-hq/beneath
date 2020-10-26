package stream

func derefBool(val *bool, fallback bool) bool {
	if val != nil {
		return *val
	}
	return fallback
}
