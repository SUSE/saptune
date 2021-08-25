package system

import (
	"math"
)

// MaxI64 returns the maximum among the input values.
// If there isn't any input value, return 0.
func MaxI64(values ...int64) int64 {
	var ret int64
	for _, value := range values {
		if ret < value {
			ret = value
		}
	}
	return ret
}

// MaxI returns the maximum among the input values.
// If there isn't any input value, return 0.
func MaxI(values ...int) int {
	var ret int
	for _, value := range values {
		if ret < value {
			ret = value
		}
	}
	return ret
}

// MaxU64 returns the maximum among the input values.
// If there isn't any input value, return 0.
func MaxU64(values ...uint64) uint64 {
	var ret uint64
	for _, value := range values {
		if ret < value {
			ret = value
		}
	}
	return ret
}

// MinU64 returns the minimum among the input values.
// If there isn't any input value, return 0.
func MinU64(values ...uint64) uint64 {
	if len(values) == 0 {
		return 0
	}
	var ret uint64 = math.MaxUint64
	for _, value := range values {
		if ret > value {
			ret = value
		}
	}
	return ret
}
