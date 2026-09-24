package valueexecution

import "testing"

func TestObserveSelfImprovementCycleFailsClosed(t *testing.T) {
	cycle := ObserveSelfImprovementCycle(
		SelfImprovementContinuationObservation{},
		SelfImprovementExecutionRequestObservation{},
	)
	if cycle.Status != SelfImprovementCycleStatusUnknown {
		t.Fatalf("status=%q, want unknown", cycle.Status)
	}
	if cycle.Reason != "CYCLE_EVIDENCE_UNTRUSTED" {
		t.Fatalf("reason=%q, want untrusted evidence", cycle.Reason)
	}
	if !cycle.NonAuthorizing || cycle.Digest == "" {
		t.Fatal("cycle must remain non-authorizing and digestable")
	}
}
