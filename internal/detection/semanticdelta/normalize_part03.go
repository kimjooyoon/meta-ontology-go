package semanticdelta

import (
	"fmt"
	"sort"
)

// IsEmpty reports whether the delta changes no semantic content.
func (d Delta) IsEmpty() bool {
	return len(d.AddedNodes) == 0 && len(d.RemovedNodes) == 0 && len(d.AddedFacts) == 0 && len(d.RemovedFacts) == 0
}
func normalizeNodes(nodes []Node, label string) ([]Node, error) {
	if len(nodes) == 0 {
		return nil, nil
	}
	result := make([]Node, len(nodes))
	seen := make(map[string]string, len(nodes))
	for i, node := range nodes {
		id, err := normalizeToken("node ID", node.ID)
		if err != nil {
			return nil, fmt.Errorf("%s[%d]: %w", label, i, err)
		}
		kind, err := normalizeToken("node kind", node.Kind)
		if err != nil {
			return nil, fmt.Errorf("%s[%d]: %w", label, i, err)
		}
		if previous, exists := seen[id]; exists && previous != kind {
			return nil, fmt.Errorf("%s: node %q has conflicting kinds %q and %q", label, id, previous, kind)
		}
		seen[id] = kind
		result[i] = Node{ID: id, Kind: kind}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].ID != result[j].ID {
			return result[i].ID < result[j].ID
		}
		return result[i].Kind < result[j].Kind
	})
	return uniqueNodes(result), nil
}
func normalizeFacts(facts []Fact, label string) ([]Fact, error) {
	if len(facts) == 0 {
		return nil, nil
	}
	result := make([]Fact, len(facts))
	for i, fact := range facts {
		subject, err := normalizeToken("fact subject", fact.Subject)
		if err != nil {
			return nil, fmt.Errorf("%s[%d]: %w", label, i, err)
		}
		predicate, err := normalizeToken("fact predicate", fact.Predicate)
		if err != nil {
			return nil, fmt.Errorf("%s[%d]: %w", label, i, err)
		}
		object, err := normalizeToken("fact object", fact.Object)
		if err != nil {
			return nil, fmt.Errorf("%s[%d]: %w", label, i, err)
		}
		result[i] = Fact{Subject: subject, Predicate: predicate, Object: object}
	}
	sort.Slice(result, func(i, j int) bool { return factLess(result[i], result[j]) })
	return uniqueFacts(result), nil
}
