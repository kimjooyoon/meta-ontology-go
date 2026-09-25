package valueexecution

import "testing"

func TestObserveSelfImprovementCounterexampleReplayRequiresRetainedEvidence(t *testing.T) {
	observation := ObserveSelfImprovementCounterexampleReplay(
		SelfImprovementCounterexampleObservation{},
		SelfImprovementCounterexampleObservation{},
	)
	if observation.Status != SelfImprovementCounterexampleReplayStatusUnknown {
		t.Fatalf("status=%q, want unknown", observation.Status)
	}
	if observation.Reason != "COUNTEREXAMPLE_NOT_RETAINED" {
		t.Fatalf("reason=%q, want not retained", observation.Reason)
	}
	if !observation.NonAuthorizing || observation.Digest == "" {
		t.Fatal("replay observation must remain non-authorizing and digestable")
	}
}
