package languagedebugexperiment

import (
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/languagedebug"
)

func testInput(t *testing.T) Input {
	t.Helper()
	first := languagedebug.Observe(executionFixture(t), "SOURCE_PARSED")
	second := languagedebug.Observe(executionFixture(t), "ACTIVITY_INVOKED")
	return Input{
		SubjectSHA: string(makeHex('a', 40)), ExecutableDigest: digest('d'),
		Contract: fixedContract(),
		First:    first, Second: second,
		UnknownBreakpoint:   languagedebug.Observe(executionFixture(t), "MISSING"),
		RuntimeObservations: []RuntimeObservation{runtimeFixture(1, first), runtimeFixture(2, second)},
		Build:               Measurement{Name: "debug-producer-build", Executed: true, WallNS: 1000000, WallMS: 1, PeakRSSKiB: 1, CacheState: "fixture"},
		EvaluatorBuild:      Measurement{Name: "debug-evaluator-build", Executed: true, WallNS: 1000000, WallMS: 1, PeakRSSKiB: 1, CacheState: "fixture"},
		Test:                Measurement{Name: "debug-relevant-tests", Executed: true, WallNS: 1000000, WallMS: 1, PeakRSSKiB: 1, CacheState: "fixture"},
		Graph: GraphObservation{Schema: "gooo-graph/v1", ProgramDigest: digest('p'), GraphHash: string(makeHex('g', 64)), ActivityCount: 44, EdgeCount: 88, DebugActivityCount: 2, DebugOutputCount: 2, DebugUsedEdgeCount: 2, DebugGeneratedEdgeCount: 2,
			DebugActivityIDs: []string{"languageutility://activity/observe-debugging-deterministic-replay", "languageutility://activity/observe-debugging-resource-observed"},
			DebugCausalEdges: []GraphEdge{{Relation: "used", Subject: "languageutility://activity/observe-debugging-deterministic-replay", Object: "gooo://meta/language-utility/entity/cell"}, {Relation: "wasGeneratedBy", Subject: "gooo://meta/language-utility/entity/evidence", Object: "languageutility://activity/observe-debugging-deterministic-replay"}, {Relation: "used", Subject: "languageutility://activity/observe-debugging-resource-observed", Object: "gooo://meta/language-utility/entity/cell"}, {Relation: "wasGeneratedBy", Subject: "gooo://meta/language-utility/entity/evidence", Object: "languageutility://activity/observe-debugging-resource-observed"}}},
	}
}

func runtimeFixture(run int, receipt languagedebug.Receipt) RuntimeObservation {
	return RuntimeObservation{Run: run, RuntimeReceiptSchema: RuntimeReceiptSchema, Runner: "fixture", Toolchain: "go version go1.27.0 fixture",
		SourceRawDigest: receipt.SourceDigest, SourceSemanticDigest: receipt.SemanticDigest,
		BinaryDigest: digest('d'), Arguments: []string{"debug"}, SubjectSHA: string(makeHex('a', 40)),
		OutputDigest: digest('o'), WallNS: 1000000, WallMS: 1, PeakRSSKiB: 1}
}

func executionFixture(t *testing.T) []byte {
	t.Helper()
	value := map[string]any{
		"schema": "gooo/source-execution-receipt/v1", "decision": "PASS", "resolution": "EXACT",
		"filename": "main.gooo", "source_digest": digest('a'), "semantic_digest": digest('b'),
		"entry": map[string]any{"activity": "PayOrder"}, "digest": digest('c'),
		"events": []map[string]any{{"sequence": 1, "kind": "SOURCE_PARSED", "subject": "a"},
			{"sequence": 2, "kind": "SEMANTIC_LOWERED", "subject": "b"},
			{"sequence": 3, "kind": "ACTIVITY_INVOKED", "subject": "PayOrder"},
			{"sequence": 4, "kind": "ENTITY_PRODUCED", "subject": "Receipt"}},
		"diagnostics": []any{}, "effects": map[string]any{"repository_writes": 0, "mutation_authority": false},
	}
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func fixedContract() Contract {
	return Contract{"gooo/language-debug-experiment-contract/v1", 1, 2, 2, 2, 4, 2, 1, 2, 4, 2, 2, 1, 3}
}

func digest(value byte) string { return "sha256:" + string(makeHex(value, 64)) }
func makeHex(value byte, size int) []byte {
	data := make([]byte, size)
	for index := range data {
		data[index] = value
	}
	return data
}
