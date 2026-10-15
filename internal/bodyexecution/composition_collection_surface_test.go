package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestCollectionSourceOrderAndOrigins7139(t *testing.T) {
	const sum = `let 합계 = 0; for 항목 in input.amounts { if 항목 > 7 { 합계 = 합계 + 항목 } }; return 합계`
	const join = `let 결과 = ""; for 이름 in input.labels { 결과 = 결과 + 이름 }; return 결과`
	source := `package collection7139
namespace collection7139
entity Integer id "urn:gooo:type:integer"
entity Text id "urn:gooo:type:string"
entity Amounts id "urn:collection7139:amounts" fields {
 field amounts id "urn:collection7139:values" type integer required many
}
entity Labels id "urn:collection7139:labels" fields {
 field labels id "urn:collection7139:names" type string required many
}
activity Sum(Amounts) -> Integer id "urn:collection7139:sum" computes ` + strconv.Quote(sum) + `
activity Join(Labels) -> Text id "urn:collection7139:join" computes ` + strconv.Quote(join) + "\n"
	const cases = `{"schema":"gooo/body-composition-cases/v1","cases":[
{"inputs":{"Sum":{"amounts":[2,19,-3,7]},"Join":{"labels":["가","a","나"]}},"expected":{"Sum":19,"Join":"가a나"}},
{"inputs":{"Sum":{"amounts":[]},"Join":{"labels":[]}},"expected":{"Sum":0,"Join":""}},
{"inputs":{"Sum":{"amounts":[9007199254740995,5]},"Join":{"labels":["나","a","가"]}},"expected":{"Sum":9007199254740995,"Join":"나a가"}}
]}`
	runCollectionSourceOrigins(t, "collection7139", source, cases, sum, "항목 > 7")
}

func TestCollectionSourceEntries7193(t *testing.T) {
	const sum = `let 합계 = 0; for 항목 in input.amounts { if 항목 >= 11 { 합계 = 합계 + 항목 } }; return 합계`
	const join = `let 결과 = ""; for 이름 in input.labels { 결과 = 결과 + 이름 }; return 결과`
	source := `package collection7193
namespace collection7193
entity Integer id "urn:gooo:type:integer"
entity Text id "urn:gooo:type:string"
entity Amounts id "urn:collection7193:amounts" fields {
 field amounts id "urn:collection7193:values" type integer required many
}
entity Labels id "urn:collection7193:labels" fields {
 field labels id "urn:collection7193:names" type string required many
}
activity Sum(Amounts) -> Integer id "urn:collection7193:sum" computes ` + strconv.Quote(sum) + `
activity Join(Labels) -> Text id "urn:collection7193:join" computes ` + strconv.Quote(join) + "\n"
	const cases = `{"schema":"gooo/body-composition-cases/v1","cases":[
{"inputs":{"Sum":{"amounts":[4,17,-2,11]},"Join":{"labels":["다","/","b"]}},"expected":{"Sum":28,"Join":"다/b"}},
{"inputs":{"Sum":{"amounts":[]},"Join":{"labels":[]}},"expected":{"Sum":0,"Join":""}},
{"inputs":{"Sum":{"amounts":[9007199254740997,6]},"Join":{"labels":["b","/","다"]}},"expected":{"Sum":9007199254740997,"Join":"b/다"}}
]}`
	runCollectionSourceOrigins(t, "collection7193", source, cases, sum, "항목 >= 11")
}

func runCollectionSourceOrigins(t *testing.T, name, source, cases, sum, expression string) {
	t.Helper()
	suite, err := DecodeCompositionCases([]byte(cases))
	if err != nil {
		t.Fatal(err)
	}
	generations := map[string]Composition{}
	runtimes := map[string]CompositionRuntime{}
	for _, entry := range []string{"Sum", "Join"} {
		selected := CompositionCases{Schema: suite.Schema}
		for _, sample := range suite.Cases {
			selected.Cases = append(selected.Cases, CompositionCase{
				Inputs:   map[string]json.RawMessage{entry: sample.Inputs[entry]},
				Expected: map[string]json.RawMessage{entry: sample.Expected[entry]},
			})
		}
		generated, err := GenerateCompositionWithOptions(context.Background(), name+".gooo", []byte(source), selected, CompositionOptions{EntryActivity: entry})
		if err != nil {
			t.Fatal("generate for-in entry", entry, err)
		}
		run, err := ExecuteComposition(context.Background(), name+".gooo", []byte(source), generated, selected, nativeTool())
		if err != nil || run.FinitePassed != 3 || run.FiniteTotal != 3 || run.ModelCalls != 0 {
			t.Fatal("for-in native entry results", entry, err, run.FinitePassed, run.FiniteTotal, run.ModelCalls)
		}
		generations[entry], runtimes[entry] = generated, run
	}
	generated, run := generations["Sum"], runtimes["Sum"]
	if len(run.ConditionSites) != 1 {
		t.Fatal("condition site count", len(run.ConditionSites))
	}
	anchor := run.ConditionSites[0].Source
	if anchor == nil || anchor.Status != "MAPPED" || anchor.Expression != expression ||
		sum[anchor.DecodedStart:anchor.DecodedEnd] != expression || generated.GoooSource[anchor.Start:anchor.End] != expression {
		t.Fatal("condition must retain original Korean source positions", anchor)
	}
	flow := bodycodegen.InspectBodyValueOrigins(sum, 1)
	if flow.Status != "PARTIAL" || !slices.Contains(flow.Reasons, "COLLECTION_ELEMENTS_JOINED") || slices.Contains(flow.Reasons, "SOURCE_RANGE_UNAVAILABLE") {
		t.Fatal("collection origin analysis must state its item-joining limit", flow)
	}
	nodes := map[string]bodycodegen.BodyValueOrigin{}
	var condition string
	for _, node := range flow.Nodes {
		nodes[node.ID] = node
		if node.Kind == "condition" {
			condition = node.ID
			if sum[node.Decoded.Start:node.Decoded.End] != expression {
				t.Fatal("value origin condition byte range moved", node)
			}
		}
	}
	seen := map[string]bool{}
	queue := []string{condition}
	inputFound := false
	for len(queue) != 0 {
		id := queue[0]
		queue = queue[1:]
		if seen[id] {
			continue
		}
		seen[id] = true
		node := nodes[id]
		inputFound = inputFound || node.Kind == "input_reference" && node.Name == "input"
		queue = append(queue, node.Inputs...)
	}
	if condition == "" || !inputFound {
		t.Fatal("condition must link through collection element to original input", flow)
	}
	if directory := os.Getenv("GOOO_COLLECTION_OBSERVATION"); directory != "" {
		raw, err := json.MarshalIndent(map[string]any{"source": source, "cases": json.RawMessage(cases), "generations": generations,
			"runtimes": runtimes, "body_value_origins": flow, "historical_source_executions": 0}, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, name+".json"), append(raw, '\n'), 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Log(name + ": native6/6, ordered Text and exact Integer collections; Korean condition source range and collection-input origin; model0")
}
