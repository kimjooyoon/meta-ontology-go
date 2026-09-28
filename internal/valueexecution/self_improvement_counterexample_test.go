package valueexecution

import "testing"

func TestObserveSelfImprovementCounterexampleRequiresReplayContext(t *testing.T) {
	observation := ObserveSelfImprovementCounterexample(
		"",
		"",
		ExecutedSelfImprovementObservation{},
		ExecutionOriginReceipt{},
		ExecutionOriginReceipt{},
	)
	if observation.Status != SelfImprovementCounterexampleStatusUnknown {
		t.Fatalf("status=%q, want unknown", observation.Status)
	}
	if observation.Reason != "COUNTEREXAMPLE_CONTEXT_MISSING" {
		t.Fatalf("reason=%q, want missing replay context", observation.Reason)
	}
	if !observation.NonAuthorizing || observation.Digest == "" {
		t.Fatal("counterexample must remain non-authorizing and digestable")
	}
}
