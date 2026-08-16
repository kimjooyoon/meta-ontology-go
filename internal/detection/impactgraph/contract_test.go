package impactgraph_test

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	impactgraph "github.com/kimjooyoon/meta-ontology-go/internal/detection/impactgraph"
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

func TestDecodeRejectsMalformedGraphs(t *testing.T) {
	for _, fixture := range []string{
		"duplicate-node.json",
		"duplicate-edge.json",
		"illegal-endpoint-kinds.json",
		"duplicate-json-field.json",
		"unknown-json-field.json",
		"trailing-json.json",
	} {
		t.Run(fixture, func(t *testing.T) {
			if _, err := impactgraph.Decode(fixtureBytes(t, fixture)); err == nil {
				t.Fatalf("Decode(%q) accepted malformed input", fixture)
			}
		})
	}
}

func TestCanonicalAndDigestReplayIgnoreInsertionOrder(t *testing.T) {
	first := decodeFixture(t, "positive-3of3.json")
	second := decodeFixture(t, "positive-3of3-reordered.json")

	firstCanonical := []byte(first.Canonical())
	secondCanonical := []byte(second.Canonical())
	if !bytes.Equal(firstCanonical, secondCanonical) {
		t.Fatalf("insertion order changed canonical bytes:\n%s\n---\n%s", firstCanonical, secondCanonical)
	}

	firstDigest := first.Digest()
	secondDigest := second.Digest()
	if firstDigest == "" || firstDigest != secondDigest {
		t.Fatalf("insertion order changed digest: %q vs %q", firstDigest, secondDigest)
	}
}

func decodeFixture(t *testing.T, name string) impactgraph.Graph {
	t.Helper()
	graph, err := impactgraph.Decode(fixtureBytes(t, name))
	if err != nil {
		t.Fatalf("Decode(%q): %v", name, err)
	}
	return graph
}

func fixtureBytes(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read fixture %q: %v", name, err)
	}
	return data
}

func assertResult(t *testing.T, got impactgraph.Result, want expectedResult) {
	t.Helper()
	if string(got.Decision) != want.decision {
		t.Errorf("Decision = %q, want %q", got.Decision, want.decision)
	}
	if got.FullSuiteRequired != want.fullSuiteRequired {
		t.Errorf("FullSuiteRequired = %v, want %v", got.FullSuiteRequired, want.fullSuiteRequired)
	}
	assertStringSet(t, "Required", got.Required, want.required)
	assertStringSet(t, "ExecutedRequired", got.ExecutedRequired, want.executedRequired)
	assertStringSet(t, "Missed", got.Missed, want.missed)
	assertStringSet(t, "Extra", got.Extra, want.extra)
	if got.Numerator != want.coverageNumerator {
		t.Errorf("Numerator = %d, want %d", got.Numerator, want.coverageNumerator)
	}
	if got.Denominator != want.coverageDenominator {
		t.Errorf("Denominator = %d, want %d", got.Denominator, want.coverageDenominator)
	}
}

func assertStringSet(t *testing.T, label string, got, want []string) {
	t.Helper()
	actual := append([]string(nil), got...)
	expected := append([]string(nil), want...)
	sort.Strings(actual)
	sort.Strings(expected)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("%s = %#v, want %#v", label, actual, expected)
	}
}
