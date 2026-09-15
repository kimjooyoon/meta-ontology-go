package artifactfeedback

import "testing"

func TestUnknownDecisionDoesNotHideReplayFailure(t *testing.T) {
	summary := Summary{RequiredInputs: 2, BoundInputs: 1, ReplayBoundInputs: 1}
	decision, reason, selected := feedbackDecision("UNKNOWN", "", summary)
	if decision != "FAIL_CLOSED" || reason != "FEEDBACK_REPLAY_UNBOUND" || selected != "" {
		t.Fatalf("decision tuple = %q/%q/%q", decision, reason, selected)
	}
}

func TestUnknownDecisionDoesNotHideAdditionalUnboundInput(t *testing.T) {
	summary := Summary{RequiredInputs: 2, BoundInputs: 0, ReplayBoundInputs: 2}
	decision, reason, selected := feedbackDecision("UNKNOWN", "", summary)
	if decision != "FAIL_CLOSED" || reason != "FEEDBACK_INPUT_UNBOUND" || selected != "" {
		t.Fatalf("decision tuple = %q/%q/%q", decision, reason, selected)
	}
}
