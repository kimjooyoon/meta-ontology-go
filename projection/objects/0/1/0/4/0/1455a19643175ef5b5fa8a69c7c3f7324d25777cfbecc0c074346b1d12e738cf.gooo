package impactgraph_test

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/impactgraph"
	"testing"
)

func TestEvaluateExtraExecutedObligationIsAllowed(t *testing.T) {
	graph := decodeFixture(t, "positive-3of3.json")
	result := graph.Evaluate(
		[]string{"change:billing/pay-order"},
		[]string{
			"obligation:billing/pay-order/unit",
			"obligation:billing/pay-order/integration",
			"obligation:billing/pay-order/provenance",
			"obligation:billing/unrelated/extra",
		},
	)

	assertResult(t, result, expectedResult{
		decision:            "PASS",
		required:            payOrderObligations,
		executedRequired:    payOrderObligations,
		extra:               []string{"obligation:billing/unrelated/extra"},
		coverageNumerator:   3,
		coverageDenominator: 3,
	})
}
func TestEvaluateUnknownChangeRequiresFullSuite(t *testing.T) {
	graph := decodeFixture(t, "positive-3of3.json")
	result := graph.Evaluate([]string{"change:billing/pay-orde"}, nil)

	assertResult(t, result, expectedResult{
		decision:          "UNKNOWN",
		fullSuiteRequired: true,
	})
}
func TestEvaluateZeroObligationsRequiresFullSuite(t *testing.T) {
	graph := decodeFixture(t, "zero-obligations.json")
	result := graph.Evaluate([]string{"change:billing/empty"}, nil)

	assertResult(t, result, expectedResult{
		decision:            "UNKNOWN",
		fullSuiteRequired:   true,
		coverageNumerator:   0,
		coverageDenominator: 0,
	})
}
func TestEvaluateUncoveredChangedRootFailsClosed(t *testing.T) {
	graph := decodeFixture(t, "positive-3of3.json")
	graph.Nodes = append(graph.Nodes, impactgraph.Node{ID: "change:billing/uncovered", Kind: impactgraph.NodeKindSemantic})
	result := graph.Evaluate(
		[]string{"change:billing/pay-order", "change:billing/uncovered"},
		payOrderObligations,
	)
	if result.Decision != impactgraph.UNKNOWN || !result.FullSuiteRequired || result.FailureCode != impactgraph.FailureCodeNoReachableObligations {
		t.Fatalf("uncovered root result = %#v", result)
	}
}
