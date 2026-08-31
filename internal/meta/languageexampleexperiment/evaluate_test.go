package languageexampleexperiment

import "testing"

func TestEvaluateObservesMinimalValueWithoutQualityClaim(t *testing.T) {
	report := Evaluate(validInput())
	if report.Decision != "PASS" || report.Interpretation != "MINIMAL_VALUE_OBSERVED" ||
		report.Summary.Coordinates.Satisfied != 15 || report.Summary.Coordinates.Total != 15 ||
		len(report.Views) != 3 || report.Views[0].Satisfied != 6 || report.Views[1].Satisfied != 12 ||
		report.Views[2].Satisfied != 15 || len(report.NotClaimed) != 5 {
		t.Fatalf("report=%#v", report)
	}
}

func TestEvaluateLowersUnknownArtifactDecision(t *testing.T) {
	input := validInput()
	input.Artifact.Decision = "UNRECOGNIZED"
	report := Evaluate(input)
	if report.Decision != "FAIL_CLOSED" || report.Resolution != "LOWER_RESOLUTION" ||
		report.Reason != "ARTIFACT_DECISION_UNKNOWN" || report.Summary.Unknowns != 1 {
		t.Fatalf("report=%#v", report)
	}
}

func TestEvaluateRejectsGoldenMismatch(t *testing.T) {
	input := validInput()
	input.Golden.Operation.Activity = "Other"
	report := Evaluate(input)
	if report.Decision != "FAIL_CLOSED" || report.Resolution != "EXACT" ||
		report.Reason != "ARTIFACT_GOLDEN_MISMATCH" {
		t.Fatalf("report=%#v", report)
	}
}
