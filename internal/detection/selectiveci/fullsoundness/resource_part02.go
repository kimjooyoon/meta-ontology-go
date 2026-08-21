package fullsoundness

import (
	"math"
)

func compareValue(left, right int64) int {
	difference, ok := subtractInt64(left, right)
	if !ok || difference == 0 {
		return 0
	}
	if difference < 0 {
		return -1
	}
	return 1
}
func addInt64(left, right int64) (int64, bool) {
	if right > 0 && left > math.MaxInt64-right {
		return 0, false
	}
	if right < 0 && left < math.MinInt64-right {
		return 0, false
	}
	return left + right, true
}
func subtractInt64(left, right int64) (int64, bool) {
	if right == math.MinInt64 {
		if left >= 0 {
			return 0, false
		}
		partial, ok := addInt64(left, math.MaxInt64)
		if !ok {
			return 0, false
		}
		return addInt64(partial, 1)
	}
	return addInt64(left, -right)
}
func multiplyInt64(left, right int64) (int64, bool) {
	if left == 0 || right == 0 {
		return 0, true
	}
	if left == -1 && right == math.MinInt64 || right == -1 && left == math.MinInt64 {
		return 0, false
	}
	product := left * right
	return product, product/right == left
}
