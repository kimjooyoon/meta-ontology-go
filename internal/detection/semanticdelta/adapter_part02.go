package semanticdelta

import (
	"fmt"
)

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
