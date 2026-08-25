package languagedebugexperiment

import "testing"

func TestEvaluateFixedDebugContract(t *testing.T) {
	input := testInput(t)
	report, err := Evaluate(input)
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != "PASS" || report.Summary.Coordinates.Satisfied != 12 {
		t.Fatalf("report = %#v", report)
	}
}

func TestEvaluateUnknownTopDecisionLowersResolution(t *testing.T) {
	input := testInput(t)
	input.First.Decision = "UNKNOWN"
	report, err := Evaluate(input)
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != "FAIL_CLOSED" || report.Resolution != "LOWER_RESOLUTION" ||
		report.Summary.Coordinates.Satisfied != 0 {
		t.Fatalf("report = %#v", report)
	}
}
