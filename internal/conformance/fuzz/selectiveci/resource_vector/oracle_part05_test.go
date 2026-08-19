package resourcevector

import (
	"strings"
)

func reverse(values []string) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}
func reversePressures(values []PressureRecord) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}
func relocate(value, oldRoot, newRoot string) string {
	return strings.Replace(value, oldRoot, newRoot, 1)
}
func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
