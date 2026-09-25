package coupling

import (
	"sort"
	"strings"
)

func baselineDelta(before, after []string) string {
	left, right := map[string]bool{}, map[string]bool{}
	for _, fact := range before {
		left[fact] = true
	}
	for _, fact := range after {
		right[fact] = true
	}
	var removed, added []string
	for fact := range left {
		if !right[fact] {
			removed = append(removed, fact)
		}
	}
	for fact := range right {
		if !left[fact] {
			added = append(added, fact)
		}
	}
	sort.Strings(removed)
	sort.Strings(added)
	if len(removed) == 0 && len(added) == 0 {
		return ""
	}
	var builder strings.Builder
	builder.WriteString("semantic-delta-v1\n")
	for _, fact := range removed {
		builder.WriteString("removed\t" + fact + "\n")
	}
	for _, fact := range added {
		builder.WriteString("added\t" + fact + "\n")
	}
	return builder.String()
}
