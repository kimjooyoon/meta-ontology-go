package selfimprovementloop

import "testing"

func TestNormalCaseClosesWithAnExactIntegerPair(t *testing.T) {
	report, err := Evaluate(testGraph(), testInput())
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionClosed || !report.PairMatched || len(report.Cells) != 12 {
		t.Fatalf("decision/pair/cells = %s/%t/%d", report.Decision, report.PairMatched, len(report.Cells))
	}
	if len(report.Unknowns) != 0 {
		t.Fatalf("unknowns = %#v", report.Unknowns)
	}
}

func TestMissingExactPairPreservesAllUnknownFields(t *testing.T) {
	input := testInput()
	input.Pair = ExactPair{}
	report, err := Evaluate(testGraph(), input)
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionUnknown || report.PairMatched {
		t.Fatalf("decision/pair = %s/%t", report.Decision, report.PairMatched)
	}
	if len(report.Unknowns) == 0 {
		t.Fatal("unknown state was discarded")
	}
	for _, state := range report.Unknowns {
		if state.Stage == "" || state.Step == "" || state.Reason == "" || state.UnknownClass == "" || state.NextOperation == "" || state.BlockedBy == "" {
			t.Fatalf("incomplete unknown state = %#v", state)
		}
	}
}
