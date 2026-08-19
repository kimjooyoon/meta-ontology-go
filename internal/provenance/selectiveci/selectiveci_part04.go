package selectiveci

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
)

func normalizeIDs(values []semantic.ID, label string) ([]semantic.ID, error) {
	if len(values) == 0 {
		return []semantic.ID{}, nil
	}
	out := make([]semantic.ID, 0, len(values))
	seen := make(map[semantic.ID]struct{}, len(values))
	for _, value := range values {
		id, err := normalizeID(value, label)
		if err != nil {
			return nil, err
		}
		if _, exists := seen[id]; exists {
			return nil, fmt.Errorf("duplicate %s %s", label, id)
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out, nil
}
func equalIDs(left, right []semantic.ID) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
func containsID(values []semantic.ID, value semantic.ID) bool {
	index := sort.Search(len(values), func(i int) bool { return values[i] >= value })
	return index < len(values) && values[index] == value
}
