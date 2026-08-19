package freshness

import (
	"sort"
)

func sortItems(items []Item) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Kind != items[j].Kind {
			return items[i].Kind < items[j].Kind
		}
		if items[i].ID != items[j].ID {
			return items[i].ID < items[j].ID
		}
		if items[i].Path != items[j].Path {
			return items[i].Path < items[j].Path
		}
		if items[i].State != items[j].State {
			return items[i].State < items[j].State
		}
		return items[i].Detail < items[j].Detail
	})
}
func join(values []string, separator string) string {
	if len(values) == 0 {
		return ""
	}
	result := values[0]
	for _, value := range values[1:] {
		result += separator + value
	}
	return result
}
