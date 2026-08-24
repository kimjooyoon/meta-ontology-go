package writeset

import (
	"path"
	"sort"
	"strings"
)

func normalizePaths(paths []string) ([]string, bool) {
	set := make(map[string]struct{}, len(paths))
	for _, candidate := range paths {
		candidate = strings.TrimSpace(strings.ReplaceAll(candidate, "\\", "/"))
		if candidate == "" {
			continue
		}
		cleaned := path.Clean(candidate)
		if cleaned == "." || strings.HasPrefix(cleaned, "../") || strings.HasPrefix(cleaned, "/") {
			return nil, false
		}
		set[cleaned] = struct{}{}
	}
	result := make([]string, 0, len(set))
	for candidate := range set {
		result = append(result, candidate)
	}
	sort.Strings(result)
	return result, true
}

func symmetricDifference(left, right []string) []string {
	counts := make(map[string]int, len(left)+len(right))
	for _, value := range left {
		counts[value]++
	}
	for _, value := range right {
		counts[value]--
	}
	result := make([]string, 0)
	for value, count := range counts {
		if count != 0 {
			result = append(result, value)
		}
	}
	sort.Strings(result)
	return result
}
