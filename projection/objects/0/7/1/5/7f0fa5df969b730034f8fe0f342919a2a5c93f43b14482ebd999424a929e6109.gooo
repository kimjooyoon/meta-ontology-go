package semanticdelta

import (
	"fmt"
)

func reconcileNodes(base []Node, delta Delta) ([]Node, error) {
	current := nodeMap(base)
	for _, node := range delta.RemovedNodes {
		existing, ok := current[node.ID]
		if !ok || existing != node {
			return nil, fmt.Errorf("reconcile nodes: cannot remove absent node %q", node.ID)
		}
		delete(current, node.ID)
	}
	for _, node := range delta.AddedNodes {
		if existing, ok := current[node.ID]; ok {
			return nil, fmt.Errorf("reconcile nodes: cannot add existing node %q (%s)", node.ID, existing.Kind)
		}
		current[node.ID] = node
	}
	result := make([]Node, 0, len(current))
	for _, node := range current {
		result = append(result, node)
	}
	return normalizeNodes(result, "reconciled nodes")
}
func reconcileFacts(base []Fact, delta Delta) ([]Fact, error) {
	current := factMap(base)
	for _, fact := range delta.RemovedFacts {
		key := factIdentityOf(fact)
		if _, ok := current[key]; !ok {
			return nil, fmt.Errorf("reconcile facts: cannot remove absent fact %q", factKey(fact))
		}
		delete(current, key)
	}
	for _, fact := range delta.AddedFacts {
		key := factIdentityOf(fact)
		if _, ok := current[key]; ok {
			return nil, fmt.Errorf("reconcile facts: cannot add existing fact %q", factKey(fact))
		}
		current[key] = fact
	}
	result := make([]Fact, 0, len(current))
	for _, fact := range current {
		result = append(result, fact)
	}
	return normalizeFacts(result, "reconciled facts")
}
