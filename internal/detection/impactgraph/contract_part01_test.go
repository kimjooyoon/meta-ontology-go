package impactgraph_test

import (
	"testing"
)

type expectedResult struct {
	decision            string
	fullSuiteRequired   bool
	required            []string
	executedRequired    []string
	missed              []string
	extra               []string
	coverageNumerator   int
	coverageDenominator int
}

var payOrderObligations = []string{
	"obligation:billing/pay-order/unit",
	"obligation:billing/pay-order/integration",
	"obligation:billing/pay-order/provenance",
}

func TestEvaluatePositiveThreeOfThree(t *testing.T) {
	graph := decodeFixture(t, "positive-3of3.json")
	result := graph.Evaluate(
		[]string{"change:billing/pay-order"},
		[]string{
			"obligation:billing/pay-order/unit",
			"obligation:billing/pay-order/integration",
			"obligation:billing/pay-order/provenance",
		},
	)

	assertResult(t, result, expectedResult{
		decision:            "PASS",
		required:            payOrderObligations,
		executedRequired:    payOrderObligations,
		coverageNumerator:   3,
		coverageDenominator: 3,
	})
}
func TestEvaluateOneMissedFailsClosed(t *testing.T) {
	graph := decodeFixture(t, "positive-3of3.json")
	result := graph.Evaluate(
		[]string{"change:billing/pay-order"},
		[]string{
			"obligation:billing/pay-order/unit",
			"obligation:billing/pay-order/provenance",
		},
	)

	assertResult(t, result, expectedResult{
		decision:            "FAIL_CLOSED",
		required:            payOrderObligations,
		executedRequired:    []string{payOrderObligations[0], payOrderObligations[2]},
		missed:              []string{"obligation:billing/pay-order/integration"},
		coverageNumerator:   2,
		coverageDenominator: 3,
	})
}
