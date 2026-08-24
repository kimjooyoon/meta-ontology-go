package verticalsliceclosureactivation

import "testing"

func TestActivationDenominatorAndMetaBindings(t *testing.T) {
	if cases := Denominator(); len(cases) != 4 { t.Fatalf("cases=%d", len(cases)) }
	receipt := Evaluate(EmbeddedInput(candidateSHA))
	classes, proofs := map[string]int{}, map[string]int{}
	for _, indicator := range receipt.Indicators { classes[indicator.Class]++; proofs[indicator.ProofChoice]++ }
	if len(receipt.Indicators) != 6 || classes["OUTCOME"] != 1 || classes["DRIVER"] != 2 || classes["GUARDRAIL"] != 3 ||
		proofs["FOUNDATION"] != 3 || proofs["COHERENCE"] != 2 || proofs["REGRESSION"] != 1 || len(receipt.MetaOperations) != 6 { t.Fatalf("indicators=%+v meta=%+v", receipt.Indicators, receipt.MetaOperations) }
}
