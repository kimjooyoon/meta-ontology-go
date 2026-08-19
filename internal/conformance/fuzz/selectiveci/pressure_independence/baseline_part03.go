package pressureindependence

import (
	"sort"
)

func withinResearchBudget(research, baseline uint64) bool {
	if baseline == 0 {
		return false
	}
	if research <= baseline {
		return true
	}
	return research <= baseline+(baseline+3)/4
}
func equalStrings(left, right []string) bool {
	left, right = append([]string(nil), left...), append([]string(nil), right...)
	sort.Strings(left)
	sort.Strings(right)
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
