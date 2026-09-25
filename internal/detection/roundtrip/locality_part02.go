package roundtrip

import (
	"slices"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func regionIDs(left, right map[semantic.ID]generatedRegion) []semantic.ID {
	values := make(map[semantic.ID]struct{}, len(left)+len(right))
	for id := range left {
		values[id] = struct{}{}
	}
	for id := range right {
		values[id] = struct{}{}
	}
	result := make([]semantic.ID, 0, len(values))
	for id := range values {
		result = append(result, id)
	}
	slices.Sort(result)
	return result
}
