package mathutil

import "fmt"

// MinInt implements min for ints
func MinInt(a, b int) int {
	if a > b {
		return b
	}
	return a
}

// MinInt64 implements min for int64s
func MinInt64(a, b int64) int64 {
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

// MaxInt64 implements max for int64s
func MaxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// MinInts returns the lowest input value
func MinInts(xs ...int) int {
	if len(xs) == 0 {
		panic(fmt.Errorf("cannot compute min on empty list"))
	}

	y := xs[0]
	for i := 1; i < len(xs); i++ {
		if xs[i] < y {
			y = xs[i]
		}
	}

	return y
}
