package artifactfeedback

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/semanticresolution"
)

func TestMissingDecisionDoesNotDescend(t *testing.T) {
	report, err := EvaluateWithResolution(ResolutionInput{
		Feedback: feedbackFixture(""),
		CurrentResolution: semanticresolution.ResolutionExactOperation,
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != "FAIL_CLOSED" ||
		report.ToResolution != report.FromResolution ||
		report.Descents != report.PreviousDescents {
		t.Fatalf("missing decision report = %#v", report)
	}
}
