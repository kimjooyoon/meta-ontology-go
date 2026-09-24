package valueexecution

import "testing"

func TestObserveSelfImprovementExecutionRequestFailsClosed(t *testing.T) {
	observation := ObserveSelfImprovementExecutionRequest(
		SelfImprovementCandidateSelectionObservation{},
		SelfImprovementExecutionRequestContext{},
	)
	if observation.Status != SelfImprovementExecutionRequestStatusUnknown {
		t.Fatalf("status=%q, want unknown", observation.Status)
	}
	if observation.Reason != "EXECUTION_REQUEST_SELECTION_UNTRUSTED" {
		t.Fatalf("reason=%q, want untrusted selection", observation.Reason)
	}
	if !observation.NonAuthorizing || observation.Digest == "" {
		t.Fatal("execution request must remain non-authorizing and digestable")
	}
}
