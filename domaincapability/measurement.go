package domaincapability

import (
	"sort"
	"strings"
)

// MeasureFromBindings derives counts from a declared domain boundary and complete bindings.
func MeasureFromBindings(expectedCapabilityIDs []string, bindings []CapabilityBinding, observationCount int) Measurement {
	expected := normalizeIDs(expectedCapabilityIDs)
	observed := SourceCapabilityIDs(bindings)
	observedSet := make(map[string]struct{}, len(observed))
	for _, id := range observed {
		observedSet[id] = struct{}{}
	}
	unresolved := 0
	for _, id := range expected {
		if _, ok := observedSet[id]; !ok {
			unresolved++
		}
	}
	return Measurement{
		ExpectedCapabilityCount:   len(expected),
		ObservedCapabilityCount:   len(observed),
		UnresolvedCapabilityCount: unresolved,
		ObservationCount:          observationCount,
		EvidenceBound:             len(observed) > 0,
	}
}

func normalizeIDs(ids []string) []string {
	seen := make(map[string]struct{}, len(ids))
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}