package artifactfeedback

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/semanticresolution"
)

func TestExplicitFeedbackDecisionsAreNeverLowered(t *testing.T) {
	for _, decision := range []string{"FIXED_POINT", "IMPROVE"} {
		report, err := EvaluateWithResolution(ResolutionInput{
			Feedback: feedbackFixture(decision),
			CurrentResolution: semanticresolution.ResolutionExactOperation,
		})
		if err != nil {
			t.Fatal(err)
		}
		if report.Decision != decision ||
			report.ToResolution != report.FromResolution ||
			report.Descents != report.PreviousDescents {
			t.Fatalf("%s report = %#v", decision, report)
		}
	}
}

func TestWriteFailureDoesNotDescend(t *testing.T) {
	feedback := feedbackFixture("UNKNOWN")
	feedback.RepositoryWrites = 1
	report, err := EvaluateWithResolution(ResolutionInput{
		Feedback: feedback,
		CurrentResolution: semanticresolution.ResolutionExactOperation,
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != "FAIL_CLOSED" ||
		report.Reason != "FEEDBACK_WRITE_EFFECT" ||
		report.ToResolution != report.FromResolution {
		t.Fatalf("write report = %#v", report)
	}
}

func TestResolutionIndicatorsAreMetaBound(t *testing.T) {
	report, err := EvaluateWithResolution(ResolutionInput{
		Feedback: feedbackFixture("UNKNOWN"),
		CurrentResolution: semanticresolution.ResolutionExactOperation,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Indicators) != 7 || report.ReportDigest == "" {
		t.Fatalf("indicator report = %#v", report)
	}
	for _, indicator := range report.Indicators {
		if !indicator.Satisfied || indicator.MetaOperation == "" ||
			indicator.Producer != "artifactfeedback.EvaluateWithResolution" {
			t.Fatalf("indicator = %#v", indicator)
		}
	}
}
