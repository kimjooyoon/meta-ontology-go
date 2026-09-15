package couplingexplain

import (
	"sort"
	"strings"
)

func sortedStrings(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	return result
}
func joinIDs(values []string) string {
	var result strings.Builder
	for _, value := range values {
		result.WriteString(value + "\x00")
	}
	return result.String()
}
