package proofchoicejudge

import (
	"encoding/json"
	"sort"
)

func evidenceDigest(ids []string, observations map[string]value) string {
	ordered := append([]string(nil), ids...)
	sort.Strings(ordered)
	values := make([]value, 0, len(ordered))
	for _, id := range ordered {
		if item, exists := observations[id]; exists {
			values = append(values, canonicalEntry(entry{Value: item}).Value)
		}
	}
	data, _ := json.Marshal(values)
	return digestBytes(data)
}

func evidenceProvenance(ids []string, observations map[string]value) []string {
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
