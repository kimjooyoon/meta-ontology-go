package semanticdelta

import (
	"fmt"
)

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
	if err := validateFactEndpoints(base.Nodes, base.Facts, "reconciliation base"); err != nil {
		return Snapshot{}, err
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
	result, err := (Snapshot{Nodes: nodes, Facts: facts}).Normalized()
	if err != nil {
		return Snapshot{}, err
	}
	if err := validateFactEndpoints(result.Nodes, result.Facts, "reconciliation result"); err != nil {
		return Snapshot{}, err
	}
	return result, nil
}
func validateFactEndpoints(nodes []Node, facts []Fact, snapshotName string) error {
	knownNodes := make(map[string]struct{}, len(nodes))
	for _, node := range nodes {
		knownNodes[node.ID] = struct{}{}
	}
	for _, fact := range facts {
		if _, ok := knownNodes[fact.Subject]; !ok {
			return fmt.Errorf(
				"reconcile %s: fact %q references missing subject node %q",
				snapshotName,
				factKey(fact),
				fact.Subject,
			)
		}
		if _, ok := knownNodes[fact.Object]; !ok {
			return fmt.Errorf(
				"reconcile %s: fact %q references missing object node %q",
				snapshotName,
				factKey(fact),
				fact.Object,
			)
		}
	}
	return nil
}
