package mathutil

// MinInt implements min for ints
func MinInt(a, b int) int {
	if a > b {
		return b
	}
	return a
}

// MaxInt implements max for ints
func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
