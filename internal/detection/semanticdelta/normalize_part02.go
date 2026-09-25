package semanticdelta

import (
	"fmt"
)

// Normalized returns a validated, sorted, detached snapshot.
func (s Snapshot) Normalized() (Snapshot, error) {
	nodes, err := normalizeNodes(s.Nodes, "snapshot nodes")
	if err != nil {
		return Snapshot{}, err
	}
	facts, err := normalizeFacts(s.Facts, "snapshot facts")
	if err != nil {
		return Snapshot{}, err
	}
	return Snapshot{Nodes: nodes, Facts: facts}, nil
}

// Normalize validates and canonicalizes the snapshot in place.
func (s *Snapshot) Normalize() error {
	if s == nil {
		return fmt.Errorf("cannot normalize a nil snapshot")
	}
	normalized, err := s.Normalized()
	if err != nil {
		return err
	}
	*s = normalized
	return nil
}

// Normalized returns a validated, sorted, detached delta.
func (d Delta) Normalized() (Delta, error) {
	addedNodes, err := normalizeNodes(d.AddedNodes, "added nodes")
	if err != nil {
		return Delta{}, err
	}
	removedNodes, err := normalizeNodes(d.RemovedNodes, "removed nodes")
	if err != nil {
		return Delta{}, err
	}
	addedFacts, err := normalizeFacts(d.AddedFacts, "added facts")
	if err != nil {
		return Delta{}, err
	}
	removedFacts, err := normalizeFacts(d.RemovedFacts, "removed facts")
	if err != nil {
		return Delta{}, err
	}
	if overlapNodes(addedNodes, removedNodes) {
		return Delta{}, fmt.Errorf("delta contains the same node in added and removed sets")
	}
	if overlapFacts(addedFacts, removedFacts) {
		return Delta{}, fmt.Errorf("delta contains the same fact in added and removed sets")
	}
	return Delta{
		AddedNodes: addedNodes, RemovedNodes: removedNodes,
		AddedFacts: addedFacts, RemovedFacts: removedFacts,
	}, nil
}

// Normalize validates and canonicalizes the delta in place.
func (d *Delta) Normalize() error {
	if d == nil {
		return fmt.Errorf("cannot normalize a nil delta")
	}
	normalized, err := d.Normalized()
	if err != nil {
		return err
	}
	*d = normalized
	return nil
}
