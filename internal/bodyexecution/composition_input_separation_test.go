package bodyexecution

import (
	"context"
	"encoding/json"
	"maps"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestCompositionInputSeparationUsesFreshNativeCases(t *testing.T) {
	source, suite, prior := retainedFieldFixture(t)
	executor := NewExecutor()
	defer executor.Close()
	run := func(cases CompositionCases) CompositionInputSeparation {
		result, err := executor.ExecuteComposition(context.Background(), "fields.gooo", source, prior, cases, nativeTool())
		if err != nil || !result.RuntimeReplayed {
			t.Fatalf("native composition: %v", err)
		}
		return result.InputSeparation
	}
	full := run(suite)
	if full.Status != "PASS" || full.UniqueInputs != 7 || full.OverlappingInputs != 5 || full.DisjointInputs != 2 || full.DisjointCasesPassed != 2 {
		t.Fatal("record source case separation differs", full)
	}
	onlySelection := suite
	onlySelection.Cases = suite.Cases[:5]
	if got := run(onlySelection); got.Status != "UNKNOWN" || got.DisjointInputs != 0 || got.OverlappingInputs != 5 {
		t.Fatal("selection inputs gained independent coverage", got)
	}
	duplicate := suite
	duplicate.Cases = append(append([]CompositionCase(nil), suite.Cases...), suite.Cases[5])
	if got := run(duplicate); got.Status != "PASS" || got.DuplicateRows != 1 || got.DisjointInputs != 2 {
		t.Fatal("duplicate input inflated coverage", got)
	}
	duplicate.Cases[7].Expected = maps.Clone(duplicate.Cases[7].Expected)
	duplicate.Cases[7].Expected["Label"] = json.RawMessage(`"conflicting expectation"`)
	if got := run(duplicate); got.Status != "PROGRESS" || got.DisjointCasesPassed != 1 || got.DisjointInputs != 2 {
		t.Fatal("conflicting duplicate expectation was discarded", got)
	}
	zero := suite
	zero.Cases = append([]CompositionCase(nil), suite.Cases[5:]...)
	for i := range zero.Cases {
		zero.Cases[i].Expected = map[string]json.RawMessage{"Label": json.RawMessage(`"not the native output"`)}
	}
	if got := run(zero); got.Status != "PROGRESS" || got.DisjointCasesPassed != 0 || got.DisjointInputs != 2 {
		t.Fatal("observed zero became unknown", got)
	}
}

func TestCompositionInputSeparationFollowsIntermediateValues(t *testing.T) {
	graph := compositionGraph{count: 2}
	graph.nodes[0] = CompositionActivity{ID: "first", InputType: "Integer"}
	graph.nodes[1] = CompositionActivity{ID: "second", InputType: "Integer", Assembling: true}
	seen := compositionSelectionInputs{1: map[string]struct{}{"[7]": {}}}
	trace := CompositionTrace{Deliveries: []CompositionDelivery{
		{ActivityID: "first", Input: json.RawMessage(`9007199254740993`), Actual: json.RawMessage(`7`)},
		{ActivityID: "second", Input: json.RawMessage(`7`), Actual: json.RawMessage(`8`)},
	}}
	if got := graph.classifyInputTrace(seen, trace); got.kind != "overlap" {
		t.Fatal("new root input hid a previously observed downstream input", got)
	}
	trace.Deliveries[1].Input = json.RawMessage(`8`)
	if got := graph.classifyInputTrace(seen, trace); got.kind != "disjoint" {
		t.Fatal("new downstream input was lost", got)
	}
	graph.nodes[1].Assembling = false
	if got := graph.classifyInputTrace(seen, trace); got.kind != "unknown" {
		t.Fatal("a plain body invented a selection contract", got)
	}
}

func TestCompositionSelectionKeysPreserveOptionalValuesAndIntegers(t *testing.T) {
	graph := compositionGraph{plan: CompositionPlan{Records: []bodycodegen.RecordType{{Name: "Value", Fields: []bodycodegen.RecordField{
		{Name: "n", TypeID: "urn:gooo:type:integer", Presence: "optional"},
		{Name: "s", TypeID: "urn:gooo:type:string"},
	}}}}}
	node := CompositionActivity{InputType: "Value"}
	keys := map[string]bool{}
	for _, raw := range []string{`{"s":""}`, `{"n":0,"s":""}`, `{"n":9007199254740992,"s":""}`, `{"n":9007199254740993,"s":""}`} {
		key, err := graph.selectionInputKey(node, []json.RawMessage{json.RawMessage(raw)})
		if err != nil || keys[key] {
			t.Fatalf("distinct typed input collapsed: %s %v", raw, err)
		}
		keys[key] = true
	}
	key, err := graph.selectionInputKey(node, []json.RawMessage{json.RawMessage(`{ "s": "", "n": 0 }`)})
	if err != nil || !keys[key] {
		t.Fatal("equivalent field order changed input identity", key, err)
	}
}
