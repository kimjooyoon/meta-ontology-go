package selfimprovementloop

import (
	"strings"
	"testing"
)

func testGraph() Graph {
	nodes := make([]GraphNode, 0, len(fixedCells))
	for _, cell := range fixedCells {
		nodes = append(nodes, GraphNode{
			ID: "gooo://meta/activity/" + strings.ToLower(cell), Kind: "Activity",
			Namespace: "meta", Name: cell,
		})
	}
	return Graph{
		SchemaVersion: GraphSchemaVersion,
		GraphHash:     strings.Repeat("1", 64),
		SourceDigest:  strings.Repeat("2", 64),
		Nodes:         nodes,
	}
}

func testInput() Input {
	source, toolchain := strings.Repeat("a", 64), strings.Repeat("b", 64)
	return Input{
		Schema: Schema, Scenario: "test", SourceDigest: source, ToolchainDigest: toolchain,
		Baseline:       BaselineObservation{Present: true, Metric: "m", Value: 1},
		Target:         TargetDeclaration{Present: true, Metric: "m", Value: 2},
		Scope:          ScopePin{Paths: []string{"internal/meta/selfimprovementloop"}},
		Transformation: TransformationProposal{Present: true, Patch: "proposal", OutputMode: "caller-owned-temporary-output"},
		Prediction:     EffectPrediction{Present: true, Metric: "m", Before: 1, After: 2},
		Counterexample: CounterexampleResult{Checked: true},
		CI:             CIResult{Executed: true, Passed: true},
		Receipt:        ReceiptInput{Captured: true, Digest: strings.Repeat("c", 64)},
		Pair: ExactPair{
			Before: []MetricSample{{Scenario: "test", SourceDigest: source, ToolchainDigest: toolchain, Metric: "m", Value: 1}},
			After:  []MetricSample{{Scenario: "test", SourceDigest: source, ToolchainDigest: toolchain, Metric: "m", Value: 2}},
		},
		Human: HumanDecision{Decision: "APPROVE"},
	}
}

func TestReleasedGraphBindsExactlyOneActivityPerCell(t *testing.T) {
	bindings, err := BindActivities(testGraph())
	if err != nil {
		t.Fatal(err)
	}
	if len(bindings) != 12 || len(SemanticCells()) != 12 {
		t.Fatalf("bindings/cells = %d/%d, want 12/12", len(bindings), len(SemanticCells()))
	}
	for index, binding := range bindings {
		if binding.Cell != fixedCells[index] || binding.Activity != fixedCells[index] || binding.ActivityID == "" {
			t.Fatalf("binding %d = %#v", index, binding)
		}
	}
}

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
