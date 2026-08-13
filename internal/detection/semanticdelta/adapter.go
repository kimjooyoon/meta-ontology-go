package semanticdelta

import "fmt"

// Adapter maps an arbitrary semantic IR representation into this package's
// small boundary model. The callbacks are the only integration point: a
// caller can adapt internal/semantic.IR without making this package import it.
// The Facts callback should return deterministic facts only; candidate facts
// remain evidence and are not semantic delta members.
type Adapter[S any] struct {
	Nodes func(S) ([]Node, error)
	Facts func(S) ([]Fact, error)
}

// Snapshot adapts one source value and validates its semantic projection.
func (a Adapter[S]) Snapshot(source S) (Snapshot, error) {
	if a.Nodes == nil || a.Facts == nil {
		return Snapshot{}, fmt.Errorf("semanticdelta adapter requires node and fact callbacks")
	}
	nodes, err := a.Nodes(source)
	if err != nil {
		return Snapshot{}, fmt.Errorf("adapt nodes: %w", err)
	}
	facts, err := a.Facts(source)
	if err != nil {
		return Snapshot{}, fmt.Errorf("adapt facts: %w", err)
	}
	return (Snapshot{Nodes: nodes, Facts: facts}).Normalized()
}

// Diff adapts two source values and computes their presentation-insensitive
// semantic delta.
func (a Adapter[S]) Diff(before, after S) (Delta, error) {
	left, right, err := a.snapshots(before, after)
	if err != nil {
		return Delta{}, err
	}
	return DiffSnapshots(left, right)
}

func (a Adapter[S]) snapshots(before, after S) (Snapshot, Snapshot, error) {
	left, err := a.Snapshot(before)
	if err != nil {
		return Snapshot{}, Snapshot{}, fmt.Errorf("adapt before snapshot: %w", err)
	}
	right, err := a.Snapshot(after)
	if err != nil {
		return Snapshot{}, Snapshot{}, fmt.Errorf("adapt after snapshot: %w", err)
	}
	return left, right, nil
}

// Apply computes and scope-checks a deterministic delta before invoking the
// caller's authoritative writer. Out-of-scope deltas and empty deltas never
// reach commit; the latter keeps candidate-only changes and replays no-op.
func (a Adapter[S]) Apply(before, after S, scope Scope, commit func(Delta) error) (Report, error) {
	left, right, err := a.snapshots(before, after)
	if err != nil {
		return Report{}, fmt.Errorf("compute semantic delta: %w", err)
	}
	delta, err := DiffSnapshots(left, right)
	if err != nil {
		return Report{}, fmt.Errorf("compute semantic delta: %w", err)
	}
	report, err := Detect(delta, scope)
	if err != nil {
		return report, err
	}
	if !report.Passes() {
		return report, &ScopeError{Report: report}
	}
	if _, err := Reconcile(left, delta); err != nil {
		return report, fmt.Errorf("reconcile semantic delta: %w", err)
	}
	if delta.IsEmpty() {
		return report, nil
	}
	if commit == nil {
		return Report{}, fmt.Errorf("semanticdelta commit callback is required")
	}
	if err := commit(delta); err != nil {
		return report, fmt.Errorf("commit semantic delta: %w", err)
	}
	return report, nil
}

// Adapt is a convenience form for callers that only need one snapshot.
func Adapt[S any](source S, adapter Adapter[S]) (Snapshot, error) {
	return adapter.Snapshot(source)
}

// DiffSnapshots computes the semantic delta between two adapter-neutral
// snapshots. Node identity includes the immutable kind; a kind change is a
// remove followed by an add for the same stable ID.
func DiffSnapshots(before, after Snapshot) (Delta, error) {
	left, err := before.Normalized()
	if err != nil {
		return Delta{}, fmt.Errorf("normalize before snapshot: %w", err)
	}
	right, err := after.Normalized()
	if err != nil {
		return Delta{}, fmt.Errorf("normalize after snapshot: %w", err)
	}
	delta := Delta{}
	leftNodes := nodeMap(left.Nodes)
	rightNodes := nodeMap(right.Nodes)
	for id, node := range leftNodes {
		if other, exists := rightNodes[id]; !exists || other != node {
			delta.RemovedNodes = append(delta.RemovedNodes, node)
		}
	}
	for id, node := range rightNodes {
		if other, exists := leftNodes[id]; !exists || other != node {
			delta.AddedNodes = append(delta.AddedNodes, node)
		}
	}
	leftFacts := factMap(left.Facts)
	rightFacts := factMap(right.Facts)
	for key, fact := range leftFacts {
		if _, exists := rightFacts[key]; !exists {
			delta.RemovedFacts = append(delta.RemovedFacts, fact)
		}
	}
	for key, fact := range rightFacts {
		if _, exists := leftFacts[key]; !exists {
			delta.AddedFacts = append(delta.AddedFacts, fact)
		}
	}
	return delta.Normalized()
}

func nodeMap(nodes []Node) map[string]Node {
	result := make(map[string]Node, len(nodes))
	for _, node := range nodes {
		result[node.ID] = node
	}
	return result
}

func factMap(facts []Fact) map[string]Fact {
	result := make(map[string]Fact, len(facts))
	for _, fact := range facts {
		result[factKey(fact)] = fact
	}
	return result
}
