package externalconformanceactivation

import "testing"

func TestDenominatorAndMetaBindings(t *testing.T) {
	if cases := Denominator(); len(cases) != 8 {
		t.Fatalf("cases=%d", len(cases))
	}
	receipt := Evaluate(EmbeddedInput(candidateSHA))
	classes, proofs := map[string]int{}, map[string]int{}
	for _, indicator := range receipt.Indicators {
		classes[indicator.Class]++
		proofs[indicator.ProofChoice]++
	}
	if len(receipt.Indicators) != 10 || classes["OUTCOME"] != 2 || classes["DRIVER"] != 4 ||
		classes["GUARDRAIL"] != 4 || proofs["FOUNDATION"] != 4 || proofs["COHERENCE"] != 2 ||
		proofs["REGRESSION"] != 4 || len(receipt.MetaOperations) != 10 {
		t.Fatalf("indicators=%+v meta=%+v", receipt.Indicators, receipt.MetaOperations)
	}
}
