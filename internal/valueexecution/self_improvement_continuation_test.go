package valueexecution

import "testing"

func TestObserveSelfImprovementContinuationFailsClosed(t *testing.T) {
	continuation := ObserveSelfImprovementContinuation(
		SelfImprovementExecutionHistoryObservation{},
		SelfImprovementCandidateSelectionObservation{},
	)
	if continuation.Status != SelfImprovementContinuationStatusUnknown {
		t.Fatalf("status=%q, want unknown", continuation.Status)
	}
	if continuation.Reason != "CONTINUATION_EVIDENCE_UNTRUSTED" {
		t.Fatalf("reason=%q, want untrusted evidence", continuation.Reason)
	}
	if !continuation.NonAuthorizing || continuation.Digest == "" {
		t.Fatal("continuation must remain non-authorizing and digestable")
	}
}
