package artifactfeedback

import "testing"

func TestFeedbackDecisionFailsClosedOnUnknownCoverageDecision(t *testing.T) {
	decision, reason, selected := feedbackDecision("UNKNOWN", "", Summary{})
	if decision != "FAIL_CLOSED" {
		t.Fatalf("decision = %q, want FAIL_CLOSED", decision)
	}
	if reason != "FEEDBACK_COVERAGE_DECISION_UNKNOWN" {
		t.Fatalf("reason = %q, want FEEDBACK_COVERAGE_DECISION_UNKNOWN", reason)
	}
	if selected != "" {
		t.Fatalf("selected operation = %q, want empty", selected)
	}
}
