package proofchoicealgebra

import (
	"encoding/json"
	"fmt"
)

type Bundle struct{ Values []Value }

// Combine is the disjoint semantic-value union. It composes evidence before
// route selection; no route is accepted merely because a producer declared it.
func Combine(left, right Bundle) (Bundle, error) {
	result := Bundle{Values: append([]Value(nil), left.Values...)}
	for _, incoming := range right.Values {
		found := -1
		for index, existing := range result.Values {
			if existing.ID == incoming.ID {
				found = index
				break
			}
		}
		if found < 0 {
			result.Values = append(result.Values, incoming)
			continue
		}
		if !sameValue(result.Values[found], incoming) {
			return Bundle{}, fmt.Errorf("PROOF_ROUTE_CONTRADICTION: %s", incoming.ID)
		}
	}
	return result, nil
}

func sameValue(left, right Value) bool {
	a, _ := json.Marshal(canonicalEntry(semanticEntry{Value: left}).Value)
	b, _ := json.Marshal(canonicalEntry(semanticEntry{Value: right}).Value)
	return string(a) == string(b)
}
