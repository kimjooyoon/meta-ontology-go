package valueexecution

import "testing"

func TestObserveSelfImprovementExecutionOutcomeFailsClosed(t *testing.T) {
	outcome := ObserveSelfImprovementExecutionOutcome(
		SelfImprovementExecutionRequestObservation{},
		ExecutionOriginReceipt{},
	)
	if outcome.Status != SelfImprovementExecutionOutcomeStatusUnknown {
		t.Fatalf("status=%q, want unknown", outcome.Status)
	}
	if outcome.Reason != "EXECUTION_OUTCOME_REQUEST_UNTRUSTED" {
		t.Fatalf("reason=%q, want untrusted request", outcome.Reason)
	}
	if !outcome.NonAuthorizing || outcome.Digest == "" {
		t.Fatal("execution outcome must remain non-authorizing and digestable")
	}
}
