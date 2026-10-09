package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestCompositionInspectionKeepsCallerPortsAndSourceIdentity(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/native-input-joins.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	got, err := InspectComposition(ctx, "joins.gooo", source, "Label")
	if err != nil {
		t.Fatal(err)
	}
	keys := []string{}
	for _, input := range got.CallerInputs {
		keys = append(keys, input.Key)
	}
	if !reflect.DeepEqual(keys, []string{"Left", "Right", "Label.input1", "Label.input2"}) {
		t.Fatal("bound input or unrelated activity became a caller input", keys)
	}
	if got.SourceSHA256 != digest(source) || got.ModelCalls != 0 || got.CandidateTests != 0 || got.NativeExecutions != 0 ||
		len(got.Assemblies) != 1 || got.Assemblies[0].Name != "Left" || got.Assemblies[0].Kind != "typed_paths" {
		t.Fatal("inspection identity or side effects", got)
	}
	again, err := InspectComposition(ctx, "another-name.gooo", source, "Label")
	if err != nil || !reflect.DeepEqual(got, again) {
		t.Fatal("inspection depends on path", err)
	}
	raw, err := CompositionInputTemplate(got)
	if err != nil {
		t.Fatal(err)
	}
	suite, err := DecodeCompositionInputs(raw)
	if err != nil {
		t.Fatal(err)
	}
	graph, err := prepareCompositionGraphForEntry(ctx, "joins.gooo", source, "Label")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = graph.inputRows(suite); err != nil {
		t.Fatal("template does not satisfy root types", err)
	}
	if strings.Contains(string(raw), "expected") {
		t.Fatal("template invented an oracle")
	}
	prior, err := GenerateCompositionWithOptions(ctx, "joins.gooo", source, suite, CompositionOptions{EntryActivity: "Label"})
	if err != nil || !reflect.DeepEqual(got.Plan, prior.Plan) {
		t.Fatal("inspection and generation plans differ", err)
	}
}

func TestCompositionInspectionShowsCalledAssemblyOnceInDependencyOrder(t *testing.T) {
	source := []byte(calledRecordSeed + `
entity Text id "called://text"
activity Wrap(Integer) -> Box computes "let next = Seed(input); return Box{value: next.value}" assembling {
 choice "value" field_value at "0" alternative "next.value * 2" intent "Double."
 value_case "[1]" -> "{\"value\":4}"
 attempts "2"
}
activity Main(Box, Text) -> Box computes "return Wrap(input0.value)"
bind Seed.result -> Main.input0
`)
	got, err := InspectComposition(context.Background(), "called.gooo", source, "Main")
	if err != nil {
		t.Fatal(err)
	}
	if len(got.CallerInputs) != 2 || got.CallerInputs[0].Key != "Seed" || got.CallerInputs[1].Key != "Main.input1" ||
		len(got.Assemblies) != 2 || got.Assemblies[0].Name != "Seed" || got.Assemblies[1].Name != "Wrap" ||
		got.Assemblies[0].Phase != "called_body" || !got.Assemblies[1].DependsOnAssembly {
		t.Fatal("caller roots and helper construction mixed", got)
	}
	if got.Assemblies[0].SourceCases != 2 || got.Assemblies[0].ContractSHA256 == "" {
		t.Fatal(got.Assemblies)
	}
}

func TestCompositionInspectionTemplateRetainsRecordPresenceAndScalarAliases(t *testing.T) {
	source := []byte(`package sample
namespace sample
entity 정수 id "urn:gooo:type:integer"
entity Work id "sample://work" fields {
 field used id "sample://work/used" type integer required one
 field note id "sample://work/note" type string optional one
 field active id "sample://work/active" type boolean required one
}
activity Main(Work, 정수) -> 정수 computes "return input0.used + input1"
`)
	got, err := InspectComposition(context.Background(), "sample.gooo", source, "Main")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := CompositionInputTemplate(got)
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Inputs []map[string]json.RawMessage `json:"inputs"`
	}
	if err = json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	var record map[string]json.RawMessage
	if err = json.Unmarshal(doc.Inputs[0]["Main.input0"], &record); err != nil {
		t.Fatal(err)
	}
	if len(record) != 2 || string(record["active"]) != "false" || string(record["used"]) != "0" ||
		string(doc.Inputs[0]["Main.input1"]) != "0" {
		t.Fatal("optional field was invented or alias lost", string(raw))
	}
}

func TestCompositionInspectionEnumeratesMultipleAssemblyProfilesWithoutTests(t *testing.T) {
	for _, family := range []string{"fill", "search"} {
		source, err := os.ReadFile("../../examples/body-codegen/source-" + family + "-composition.gooo.fixture")
		if err != nil {
			t.Fatal(err)
		}
		got, err := InspectComposition(context.Background(), family+".gooo", source, "")
		if err != nil {
			t.Fatal(err)
		}
		if len(got.Assemblies) != 2 || got.Assemblies[0].Kind != "source_"+family ||
			got.Assemblies[1].Kind != "source_"+family || got.CandidateTests != 0 {
			t.Fatal("assembly profiles collapsed", got)
		}
	}
	// Structural inspection deliberately does not certify a body's return type.
	source := []byte(`package structural
namespace structural
entity Integer id "structural://integer"
activity Main(Integer) -> Integer computes "return true"
`)
	got, err := InspectComposition(context.Background(), "structural.gooo", source, "Main")
	if err != nil || len(got.Assemblies) != 0 || got.CandidateTests != 0 {
		t.Fatal(err, got)
	}
	suite := calledCompositionCases(t, `{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"Main":0},"expected":{"Main":0}}]}`)
	if _, err = GenerateComposition(context.Background(), "structural.gooo", source, suite, ""); err == nil {
		t.Fatal("structural metadata accidentally certified a wrong body")
	}
}

func TestCompositionInspectionRejectsInvalidContextEntryAndCycles(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, active := range []context.Context{nil, ctx} {
		if _, err := InspectComposition(active, "bad.gooo", []byte(calledRecordSeed), "Seed"); err == nil {
			t.Fatal("inactive context accepted")
		}
	}
	for _, source := range []string{calledRecordSeed, calledRecordSeed + `activity Main(Integer) -> Box computes "return Main(input)"`} {
		if _, err := InspectComposition(context.Background(), "bad.gooo", []byte(source), "Main"); err == nil {
			t.Fatal("invalid graph accepted")
		}
	}
}

func TestCompositionInspectionTemplateRejectsAmbiguousOrOversizedPlans(t *testing.T) {
	base := CompositionInspection{Schema: "gooo/body-composition-inspection/v1", CallerInputs: []CompositionCallerInput{
		{Key: "Main", ScalarKind: "Integer"},
	}}
	for _, change := range []func(*CompositionInspection){
		func(p *CompositionInspection) { p.Schema = "unknown" },
		func(p *CompositionInspection) { p.CallerInputs = nil },
		func(p *CompositionInspection) { p.CallerInputs = append(p.CallerInputs, p.CallerInputs[0]) },
		func(p *CompositionInspection) { p.CallerInputs[0].Key = "" },
		func(p *CompositionInspection) { p.CallerInputs[0].ScalarKind = "unknown" },
		func(p *CompositionInspection) { p.CallerInputs[0].Key = strings.Repeat("x", 32<<10) },
	} {
		p := base
		p.CallerInputs = append([]CompositionCallerInput(nil), base.CallerInputs...)
		change(&p)
		if _, err := CompositionInputTemplate(p); err == nil {
			t.Fatal("invalid template accepted", p.Schema)
		}
	}
}
