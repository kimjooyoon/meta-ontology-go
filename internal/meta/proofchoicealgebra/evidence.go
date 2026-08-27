package proofchoicealgebra

import (
	"encoding/json"
	"sort"
)

func evidenceDigest(ids []string, observations map[string]Value) string {
	ordered := append([]string(nil), ids...)
	sort.Strings(ordered)
	values := make([]Value, 0, len(ordered))
	for _, id := range ordered {
		if value, exists := observations[id]; exists {
			values = append(values, canonicalEntry(semanticEntry{Value: value}).Value)
		}
	}
	data, _ := json.Marshal(values)
	return digestBytes(data)
}

func evidenceProvenance(ids []string, observations map[string]Value) []string {
	seen := map[string]bool{}
	result := []string{}
	for _, id := range ids {
		for _, provenance := range observations[id].Provenance {
			if provenance != "" && !seen[provenance] {
				seen[provenance] = true
				result = append(result, provenance)
			}
		}
	}
	sort.Strings(result)
	return result
}
