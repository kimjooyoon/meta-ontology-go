package query

func hasRepeatedID(ids []ID) bool {
	seen := make(map[ID]struct{}, len(ids))
	for _, id := range ids {
		if _, exists := seen[id]; exists {
			return true
		}
		seen[id] = struct{}{}
	}
	return false
}
