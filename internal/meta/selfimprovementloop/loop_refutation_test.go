package selfimprovementloop

import "testing"

func TestRefutedCaseTakesPriorityOverUnknown(t *testing.T) {
	input := testInput()
	input.Pair = ExactPair{}
	input.Counterexample = CounterexampleResult{Checked: true, Found: true, Evidence: "disproof"}
	report, err := Evaluate(testGraph(), input)
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionRefuted || report.Reason != "REFUTED_TAKES_PRIORITY_OVER_UNKNOWN" {
		t.Fatalf("decision/reason = %s/%s", report.Decision, report.Reason)
	}
	if len(report.Unknowns) == 0 {
		t.Fatal("unknown evidence was not retained alongside refutation")
	}
}

func TestMismatchedPairContextIsRefuted(t *testing.T) {
	input := testInput()
	input.Pair.Before[0].Scenario = "other"
	report, err := Evaluate(testGraph(), input)
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionRefuted {
		t.Fatalf("decision = %s, want REFUTED", report.Decision)
	}
}
