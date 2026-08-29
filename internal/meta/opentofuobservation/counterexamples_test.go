package opentofuobservation

import "testing"

func TestFixedCounterexamplesHaveTypedUnknowns(t *testing.T) {
	items := FixedCounterexamples()
	if len(items) != 9 {
		t.Fatalf("counterexamples=%d", len(items))
	}
	for _, item := range items {
		if item.Expected == DecisionUnknown {
			if item.Unknown == nil || item.Unknown.Stage == "" || item.Unknown.Step == "" || item.Unknown.Reason == "" || item.Unknown.UnknownClass == "" || item.Unknown.NextOperation == "" || item.Unknown.BlockedBy == nil {
				t.Fatalf("unknown case %q is not six-field typed", item.ID)
			}
		}
	}
}

func TestFixedCellsHaveExactProofAndIndicatorDenominators(t *testing.T) {
	if len(fixedCells) != 12 { t.Fatal("cell denominator drift") }
	proofs, indicators := map[string]int{}, map[string]int{}
	for _, cell := range fixedCells { proofs[cell.ProofChoice]++; indicators[cell.Indicator]++ }
	for _, value := range []string{"FOUNDATION", "COHERENCE", "REGRESSION"} { if proofs[value] != 4 { t.Fatalf("proof %s=%d", value, proofs[value]) } }
	for _, value := range []string{"DRIVER", "OUTCOME", "GUARDRAIL"} { if indicators[value] != 4 { t.Fatalf("indicator %s=%d", value, indicators[value]) } }
}
