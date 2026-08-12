package semanticdelta

import "fmt"

// Reconcile applies a semantic delta to a snapshot and returns the canonical
// resulting snapshot. Inputs are normalized into detached values first, so a
// failed or successful reconciliation never mutates caller-owned slices.
// Removes are applied before adds, which permits an immutable kind change to
// be represented as a remove followed by an add for one stable ID.
func Reconcile(before Snapshot, delta Delta) (Snapshot, error) {
	base, err := before.Normalized()
	if err != nil {
		return Snapshot{}, fmt.Errorf("normalize reconciliation base: %w", err)
	}
	change, err := delta.Normalized()
	if err != nil {
		return Snapshot{}, fmt.Errorf("normalize reconciliation delta: %w", err)
	}
	nodes, err := reconcileNodes(base.Nodes, change)
	if err != nil {
		return Snapshot{}, err
	}
	facts, err := reconcileFacts(base.Facts, change)
	if err != nil {
		return Snapshot{}, err
	}
	return (Snapshot{Nodes: nodes, Facts: facts}).Normalized()
}

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
		key := factKey(fact)
		if _, ok := current[key]; !ok {
			return nil, fmt.Errorf("reconcile facts: cannot remove absent fact %q", key)
		}
		delete(current, key)
	}
	for _, fact := range delta.AddedFacts {
		key := factKey(fact)
		if _, ok := current[key]; ok {
			return nil, fmt.Errorf("reconcile facts: cannot add existing fact %q", key)
		}
		current[key] = fact
	}
	result := make([]Fact, 0, len(current))
	for _, fact := range current {
		result = append(result, fact)
	}
	return normalizeFacts(result, "reconciled facts")
}
