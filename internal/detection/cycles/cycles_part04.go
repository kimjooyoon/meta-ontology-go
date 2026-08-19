package cycles

func joinIDs(ids []ID) string {
	result := ""
	for i, id := range ids {
		if i > 0 {
			result += " -> "
		}
		result += id
	}
	return result
}
