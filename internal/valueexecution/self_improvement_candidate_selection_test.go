package valueexecution

import "testing"

func TestObserveSelfImprovementCandidateSelectionFailsClosed(t *testing.T) {
	selection := ObserveSelfImprovementCandidateSelection(
		SelfImprovementPromotionObservation{},
		SelfImprovementCounterexampleReplayObservation{},
	)
	if selection.Status != SelfImprovementCandidateSelectionStatusUnknown {
		t.Fatalf("status=%q, want unknown", selection.Status)
	}
	if selection.Reason != "CANDIDATE_SELECTION_EVIDENCE_UNTRUSTED" {
		t.Fatalf("reason=%q, want untrusted evidence", selection.Reason)
	}
	if !selection.NonAuthorizing || selection.Digest == "" {
		t.Fatal("candidate selection must remain non-authorizing and digestable")
	}
}
