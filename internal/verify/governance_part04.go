package verify

import (
	"slices"
	"sort"
)

func sameStrings(left, right []string) bool {
	left = append([]string(nil), left...)
	right = append([]string(nil), right...)
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
func contains(values []string, want string) bool {
	return slices.Contains(values, want)
}
func canonicalJobs() []string {
	return []string{"gofmt", "go vet", "go test", "go test -race", "Semantic conformance", "CI policy"}
}
