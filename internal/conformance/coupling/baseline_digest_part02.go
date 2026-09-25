package coupling

import (
	"sort"
)

func baselineChangedLabels(input Input) []string {
	result := make([]string, 0)
	for _, change := range input.Changes {
		if change.BeforeDigest != change.AfterDigest {
			result = append(result, change.CodeSymbolID)
		}
	}
	sort.Strings(result)
	return result
}
func sameSurfaceSet(left, right []string) bool {
	left, right = append([]string(nil), left...), append([]string(nil), right...)
	sort.Strings(left)
	sort.Strings(right)
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
