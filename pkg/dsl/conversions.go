package dsl

import (
	"fmt"
	"math"
)

func checkedUint32FromInt(v int, name string) (uint32, error) {
	if v < 0 {
		return 0, fmt.Errorf("%s must be non-negative, got %d", name, v)
	}
	if v > math.MaxUint32 {
		return 0, fmt.Errorf("%s exceeds uint32 range: %d", name, v)
	}
	return uint32(v), nil
}

func checkedUint32FromInt64(v int64, name string) (uint32, error) {
	if v < 0 {
		return 0, fmt.Errorf("%s must be non-negative, got %d", name, v)
	}
	if v > math.MaxUint32 {
		return 0, fmt.Errorf("%s exceeds uint32 range: %d", name, v)
	}
	return uint32(v), nil
}
