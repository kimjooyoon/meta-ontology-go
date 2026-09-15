package writeset

func changedPaths(before, after Snapshot) []string {
	left := make(map[string]Entry, len(before.Entries))
	right := make(map[string]Entry, len(after.Entries))
	for _, entry := range before.Entries {
		left[entry.Path] = entry
	}
	for _, entry := range after.Entries {
		right[entry.Path] = entry
	}
	candidates := make([]string, 0, len(left)+len(right))
	for candidate := range left {
		if right[candidate] != left[candidate] {
			candidates = append(candidates, candidate)
		}
	}
	for candidate := range right {
		if _, exists := left[candidate]; !exists {
			candidates = append(candidates, candidate)
		}
	}
	result, _ := normalizePaths(candidates)
	return result
}
