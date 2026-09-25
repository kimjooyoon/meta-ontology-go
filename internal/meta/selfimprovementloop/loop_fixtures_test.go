package selfimprovementloop

import "strings"

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
