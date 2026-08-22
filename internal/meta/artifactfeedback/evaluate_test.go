package artifactfeedback

import (
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/artifactcoverage"
)

func feedbackFixture(decision string) Input {
	head := strings.Repeat("a", 40)
	coverage := artifactcoverage.Report{
		Schema: artifactcoverage.ReportSchema, CommitSHA: head,
		Repository: "kimjooyoon/meta-ontology-go", Decision: decision,
		Summary: artifactcoverage.Summary{RequiredOperations: 5, CanonicalOperations: 5},
	}
	if decision == "IMPROVE" {
		coverage.SelectedOperation = "split-go-declarations"
	}
	coverage.ReportDigest = digestJSON(coverage)
	cycle := CycleObservation{Schema: CycleSchema, HeadSHA: head, Status: "BOUND",
		CIConclusion: "success", EnvelopeDigest: strings.Repeat("1", 64), ReplayDigest: strings.Repeat("2", 64)}
	return Input{Coverage: coverage, CoverageReplayDigest: coverage.ReportDigest, Cycle: cycle}
}

func TestEvaluateFixedPointFeedback(t *testing.T) {
	report, err := Evaluate(feedbackFixture("FIXED_POINT"))
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != "FIXED_POINT" || report.Summary.ReadinessBasisPoints != 10000 || report.NextOperation != "" {
		t.Fatalf("fixed feedback = %#v", report)
	}
}

func TestEvaluateSelectsNextOperation(t *testing.T) {
	report, err := Evaluate(feedbackFixture("IMPROVE"))
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != "IMPROVE" || report.NextOperation != "split-go-declarations" {
		t.Fatalf("improvement feedback = %#v", report)
	}
}

func TestEvaluateStaleInputFailsClosed(t *testing.T) {
	input := feedbackFixture("FIXED_POINT")
	input.Cycle.HeadSHA = strings.Repeat("b", 40)
	report, err := Evaluate(input)
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != "FAIL_CLOSED" || report.Reason != "FEEDBACK_INPUT_STALE" {
		t.Fatalf("stale feedback = %#v", report)
	}
}

func TestEvaluateWriteEffectFailsClosed(t *testing.T) {
	input := feedbackFixture("FIXED_POINT")
	input.RepositoryWrites = 1
	report, err := Evaluate(input)
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != "FAIL_CLOSED" || report.Reason != "FEEDBACK_WRITE_EFFECT" {
		t.Fatalf("write feedback = %#v", report)
	}
}
