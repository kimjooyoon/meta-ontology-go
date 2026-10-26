package artifactfeedback

import "testing"

func TestCancelledCycleCannotSelectNextMetaOperation(t *testing.T) {
	for _, state := range []struct{ status, conclusion string }{
		{"INTERRUPTED", "cancelled"}, {"BOUND", "cancelled"}, {"INTERRUPTED", "success"}, {"BOUND", "failure"},
	} {
		input := feedbackFixture("IMPROVE")
		input.Cycle.Status, input.Cycle.CIConclusion = state.status, state.conclusion
		got, err := Evaluate(input)
		if err != nil || got.Decision != "FAIL_CLOSED" || got.Reason != "FEEDBACK_INPUT_UNBOUND" ||
			got.NextOperation != "" || got.Summary.ReadinessBasisPoints >= 10000 {
			t.Fatal("unfinished or failed source CI selected more work", state, got, err)
		}
	}
}
