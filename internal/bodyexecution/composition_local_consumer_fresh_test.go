package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestGoooLocalConsumer10399(t *testing.T) {
	const source = `package locals10399
namespace locals10399
entity Integer id "urn:gooo:type:integer"
entity Batch id "urn:locals10399:batch" fields {field values id "urn:locals10399:values" type integer required many}
activity Prepare(Integer) -> Batch id "urn:locals10399:prepare" computes "let first=input+163;let second=first*2;first=second-3;return Batch{values:[]int64{first,second}}"
`
	exerciseLocalConsumer(t, source, [][2]string{{`17`, `{"values":[357,360]}`}, {`19`, `{"values":[361,364]}`}}, "local-consumer10399.json")
}

func TestGoooLocalConsumer10433(t *testing.T) {
	const source = `package locals10433
namespace locals10433
entity Integer id "urn:gooo:type:integer"
entity Batch id "urn:locals10433:batch" fields {field values id "urn:locals10433:values" type integer required many}
activity Prepare(Integer) -> Batch id "urn:locals10433:prepare" computes "let first=input+173;let second=first*3;if input<0 {first=second+5}else{first=second-5};return Batch{values:[]int64{second,first}}"
`
	exerciseLocalConsumer(t, source, [][2]string{{`23`, `{"values":[588,583]}`}, {`-181`, `{"values":[-24,-19]}`}}, "local-consumer10433.json")
}

func exerciseLocalConsumer(t *testing.T, source string, inputs [][2]string, observationName string) {
	t.Helper()
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1"}
	for _, v := range inputs {
		suite.Cases = append(suite.Cases, CompositionCase{Inputs: map[string]json.RawMessage{"Prepare": json.RawMessage(v[0])}, Expected: map[string]json.RawMessage{"Prepare": json.RawMessage(v[1])}})
	}
	o := map[string]any{"source": source, "cases": suite}
	defer saveReadObservation(t, observationName, o)
	p, err := GenerateCompositionWithOptions(context.Background(), observationName+".gooo", []byte(source), suite, CompositionOptions{EntryActivity: "Prepare"})
	o["composition"] = p
	if err != nil {
		t.Fatal(err)
	}
	r, err := ExecuteComposition(context.Background(), observationName+".gooo", []byte(source), p, suite, "")
	o["producer_runtime"] = r
	if err != nil || r.FinitePassed != 2 {
		t.Fatal("new producer", err, r.FinitePassed)
	}
	rows, report, err := SelectRecordedNativeOrigins(r, NativeOriginQuery{Kind: "local-uses", Offset: 2, Limit: 3})
	o["origin_input"] = report
	o["selected_rows"] = rows
	if err != nil || report.TotalRows != 8 || report.NextOffset == nil || *report.NextOffset != 5 || report.LocalObservation == nil || report.ProducerExecutions != 0 || report.ProducerModelCalls != 0 {
		t.Fatal("read-only selected page", err, report)
	}
	for kind, total := range map[string]int{"local-writes": 6, "local-elements": 4} {
		_, page, err := SelectRecordedNativeOrigins(r, NativeOriginQuery{Kind: kind, Limit: 1})
		if err != nil || page.TotalRows != total || page.ReturnedRows != 1 {
			t.Fatal("local fact population", kind, err, page)
		}
	}
	root := os.Getenv("GOOO_FRESH_SOURCE_ROOT")
	if root == "" {
		root = "../.."
	}
	filename := filepath.Join(root, "examples/native-origin-reader/locals/locals.gooo")
	reader, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	consumerCases := CompositionCases{Schema: "gooo/body-composition-cases/v1"}
	for _, row := range rows {
		var fact CompositionLocalUseFact
		if err := json.Unmarshal(row, &fact); err != nil {
			t.Fatal(err)
		}
		if !fact.Resolved || fact.DeclarationMapping != "MAPPED" || fact.ProjectionSHA256 != r.GeneratedSHA256 || fact.DeclarationSHA256 != r.SelectedSourceSHA256 {
			t.Fatal("coordinate scopes", fact)
		}
		expected, _ := json.Marshal(map[string]any{"case_index": fact.CaseIndex, "use_index": fact.UseIndex, "assignment_index": fact.AssignmentIndex, "variable_name": fact.VariableName, "classification": "assignment"})
		consumerCases.Cases = append(consumerCases.Cases, CompositionCase{Inputs: map[string]json.RawMessage{"Explain": row}, Expected: map[string]json.RawMessage{"Explain": expected}})
	}
	consumer, err := GenerateCompositionWithOptions(context.Background(), filename, reader, consumerCases, CompositionOptions{EntryActivity: "Explain"})
	o["consumer"] = consumer
	if err != nil {
		t.Fatal("Gooo consumer generation", err)
	}
	observed, err := ExecuteComposition(context.Background(), filename, reader, consumer, consumerCases, "")
	o["consumer_runtime"] = observed
	if err != nil || observed.FinitePassed != 3 || !observed.RuntimeReplayed {
		t.Fatal("Gooo consumer", err, observed.FinitePassed)
	}
	t.Log("new producer2/2; original executions during paging0; Gooo reader3/3; local writes6/uses8/elements4; paged source and projection coordinates remain distinct")
}
