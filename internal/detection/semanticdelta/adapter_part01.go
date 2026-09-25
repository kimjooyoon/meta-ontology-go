package semanticdelta

import (
	"fmt"
)

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
