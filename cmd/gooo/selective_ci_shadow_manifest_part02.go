package main

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
)

func sortedUnion(left, right []string) []string {
	values := append(append([]string{}, left...), right...)
	sort.Strings(values)
	return uniqueStrings(values)
}
func sortedSemanticIDs(values []semantic.ID) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, value.String())
	}
	sort.Strings(result)
	return result
}
func uniqueStrings(values []string) []string {
	if len(values) < 2 {
		return values
	}
	result := values[:1]
	for _, value := range values[1:] {
		if value != result[len(result)-1] {
			result = append(result, value)
		}
	}
	return result
}
