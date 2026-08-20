package roundtrip

import (
	"slices"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
)

func affectedIDs(delta Delta, before, after semantic.IR) []semantic.ID {
	values := make(map[semantic.ID]struct{})
	for _, id := range delta.TouchedIDs {
		values[id] = struct{}{}
	}
	facts := append(before.Graph.DeterministicFacts(), after.Graph.DeterministicFacts()...)
	for _, fact := range facts {
		if containsID(delta.TouchedIDs, fact.Subject) {
			values[fact.Object] = struct{}{}
		}
		if containsID(delta.TouchedIDs, fact.Object) {
			values[fact.Subject] = struct{}{}
		}
	}
	return sortedIDs(values)
}
func changedLocality(before, after semantic.IR) []semantic.ID {
	delta, err := SemanticDelta(before, after)
	if err != nil {
		return nil
	}
	return append([]semantic.ID(nil), delta.AffectedIDs...)
}
func hasIR(left, right semantic.IR) bool {
	return len(left.Graph.Nodes()) > 0 || len(left.Graph.DeterministicFacts()) > 0 ||
		len(right.Graph.Nodes()) > 0 || len(right.Graph.DeterministicFacts()) > 0
}
func containsID(values []semantic.ID, target semantic.ID) bool {
	index := sort.Search(len(values), func(index int) bool { return values[index] >= target })
	return index < len(values) && values[index] == target
}
func sortedIDs(values map[semantic.ID]struct{}) []semantic.ID {
	result := make([]semantic.ID, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	slices.Sort(result)
	return result
}
