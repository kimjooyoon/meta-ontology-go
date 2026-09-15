package artifactresolutionexperiment

import "testing"

func TestEvaluateExactResolutionExpansion(t *testing.T) {
	report := Evaluate(validInput())
	if report.Decision != "PASS" || report.Summary.Coordinates.Satisfied != ExpectedIndicators {
		t.Fatalf("report = %#v", report)
	}
	if report.Views[0].Satisfied != 5 || report.Views[1].Satisfied != 10 ||
		report.Views[2].Satisfied != ExpectedIndicators {
		t.Fatalf("views = %#v", report.Views)
	}
}

func TestEvaluateUnknownDecisionLowersResolution(t *testing.T) {
	input := validInput()
	input.Interface.Decision = "UNKNOWN"
	report := Evaluate(input)
	if report.Decision != "FAIL_CLOSED" || report.Resolution != "LOWER_RESOLUTION" ||
		report.Summary.Coordinates.Satisfied != 0 || report.Summary.Unknowns != 1 {
		t.Fatalf("report = %#v", report)
	}
}

func TestEvaluateGoldenDriftFailsExact(t *testing.T) {
	input := validInput()
	input.InterfaceGolden.Operation.Activity = "Other"
	report := Evaluate(input)
	if report.Decision != "FAIL_CLOSED" || report.Resolution != "EXACT" ||
		report.Summary.Coordinates.Satisfied != ExpectedIndicators-1 {
		t.Fatalf("report = %#v", report)
	}
}
