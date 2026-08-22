package artifactfeedback

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/semanticresolution"
)

func TestUnknownDecisionLowersResolutionWithoutFixedPoint(t *testing.T) {
	report, err := EvaluateWithResolution(ResolutionInput{
		Feedback: feedbackFixture("UNKNOWN"),
		CurrentResolution: semanticresolution.ResolutionExactOperation,
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.Feedback.Decision != "FAIL_CLOSED" ||
		report.Feedback.Reason != ReasonCoverageDecisionUnknown {
		t.Fatalf("feedback = %#v", report.Feedback)
	}
	if report.Decision != DecisionLowerResolution ||
		report.ToResolution != semanticresolution.ResolutionOperationClass ||
		report.NextOperation != NextOperationReevaluateFeedback {
		t.Fatalf("resolution report = %#v", report)
	}
}

func TestUnknownDecisionResolutionIsFinite(t *testing.T) {
	current := semanticresolution.ResolutionExactOperation
	descents := 0
	for _, want := range []semanticresolution.Resolution{
		semanticresolution.ResolutionOperationClass,
		semanticresolution.ResolutionInvariantOnly,
	} {
		report, err := EvaluateWithResolution(ResolutionInput{
			Feedback: feedbackFixture("UNKNOWN"),
			CurrentResolution: current, Descents: descents,
		})
		if err != nil {
			t.Fatal(err)
		}
		if report.Decision != DecisionLowerResolution || report.ToResolution != want {
			t.Fatalf("resolution report = %#v", report)
		}
		current, descents = report.ToResolution, report.Descents
	}
	report, err := EvaluateWithResolution(ResolutionInput{
		Feedback: feedbackFixture("UNKNOWN"),
		CurrentResolution: current, Descents: descents,
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != "FAIL_CLOSED" ||
		report.Reason != "SEMANTIC_RESOLUTION_BUDGET_EXHAUSTED" {
		t.Fatalf("terminal report = %#v", report)
	}
}
