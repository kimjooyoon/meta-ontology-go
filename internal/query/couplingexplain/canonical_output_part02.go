package couplingexplain

import (
	"sort"
)

func sortedStrings(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	return result
}
func joinIDs(values []string) string {
	result := ""
	for _, value := range values {
		result += value + "\x00"
	}
	return result
}
