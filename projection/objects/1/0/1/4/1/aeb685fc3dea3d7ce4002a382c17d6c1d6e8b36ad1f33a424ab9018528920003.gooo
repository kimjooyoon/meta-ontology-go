package workfrontier

func hasDuplicate(values []string) bool {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if _, exists := seen[value]; exists {
			return true
		}
		seen[value] = struct{}{}
	}
	return false
}
func conflictsWithAnyPath(path RepairPath, selected []RepairPath) bool {
	for _, other := range selected {
		if conflicts(path, other) {
			return true
		}
	}
	return false
}
