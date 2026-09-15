package cycles

import "strings"

func joinIDs(ids []ID) string {
	var result strings.Builder
	for i, id := range ids {
		if i > 0 {
			result.WriteString(" -> ")
		}
		result.WriteString(id)
	}
	return result.String()
}
