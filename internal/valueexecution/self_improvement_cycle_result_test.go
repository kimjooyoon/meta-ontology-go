package valueexecution

import "testing"

func TestObserveSelfImprovementCycleResultUnknownWithoutEvidence(t *testing.T) {
	result := ObserveSelfImprovementCycleResult(
		SelfImprovementCycleObservation{},
		SelfImprovementExecutionOutcomeObservation{},
	)
	if result.Status != SelfImprovementCycleResultStatusUnknown {
		t.Fatalf("status = %q, want UNKNOWN", result.Status)
	}
	if result.Reason != "CYCLE_RESULT_EVIDENCE_UNTRUSTED" {
		t.Fatalf("reason = %q, want CYCLE_RESULT_EVIDENCE_UNTRUSTED", result.Reason)
	}
	if !result.NonAuthorizing || !result.ReverseObservation {
		t.Fatal("cycle result must remain non-authorizing and reverse-observational")
	}
	if result.Digest == "" {
		t.Fatal("cycle result digest must be present")
	}
}

func TestObserveSelfImprovementCycleResultCompleted(t *testing.T) {
	result := ObserveSelfImprovementCycleResult(
		SelfImprovementCycleObservation{
			Status:         "READY",
			NonAuthorizing: true,
			Digest:         "cycle-digest",
		},
		SelfImprovementExecutionOutcomeObservation{
			RequestDigest:  "request-digest",
			ReceiptDigest:  "receipt-digest",
			Status:         SelfImprovementExecutionOutcomeStatusCompleted,
			NonAuthorizing: true,
			Digest:         "outcome-digest",
		},
	)
	if result.Status != SelfImprovementCycleResultStatusCompleted {
		t.Fatalf("status = %q, want COMPLETED", result.Status)
	}
	if result.Reason != "CYCLE_RESULT_REVERSE_OBSERVED" {
		t.Fatalf("reason = %q, want CYCLE_RESULT_REVERSE_OBSERVED", result.Reason)
	}
	if result.EvidenceCount != 4 {
		t.Fatalf("evidence count = %d, want 4", result.EvidenceCount)
	}
}
