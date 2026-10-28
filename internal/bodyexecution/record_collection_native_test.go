package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"testing"
)

func nestedLineItemsSource(t *testing.T) []byte {
	t.Helper()
	source, err := os.ReadFile("../../examples/record-collections/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	return source
}

func TestRecordCollectionNativeTotalsAndSavedReplay(t *testing.T) {
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1"}
	for _, row := range []struct{ input, expected string }{
		{`{"items":[]}`, `0`},
		{`{"items":[{"quantity":2,"price":7}]}`, `14`},
		{`{"items":[{"quantity":2,"price":7},{"quantity":2,"price":7}]}`, `28`},
		{`{"items":[{"quantity":1,"price":9007199254740993}]}`, `9007199254740993`},
	} {
		suite.Cases = append(suite.Cases, CompositionCase{Inputs: map[string]json.RawMessage{"Total": json.RawMessage(row.input)}, Expected: map[string]json.RawMessage{"Total": json.RawMessage(row.expected)}})
	}
	prior, err := GenerateComposition(context.Background(), "totals.gooo", nestedLineItemsSource(t), suite, "")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(prior)
	if err != nil {
		t.Fatal(err)
	}
	var saved Composition
	if err := json.Unmarshal(raw, &saved); err != nil {
		t.Fatal(err)
	}
	run, err := ExecuteComposition(context.Background(), "totals.gooo", nestedLineItemsSource(t), saved, suite, nativeTool())
	if err != nil || run.FinitePassed != 4 || run.FiniteTotal != 4 || !run.RuntimeReplayed {
		t.Fatalf("record collection native/replay: %v stage=%s finite=%d/%d build=%s", err, run.Stage, run.FinitePassed, run.FiniteTotal, run.Build.Diagnostics)
	}
	items := run.Traces[2].Deliveries[0].InputFields[0]
	if items.ID != "urn:record-items:items" || len(items.Children) != 2 {
		t.Fatal("record collection container origin lost", items)
	}
	for i, child := range items.Children {
		if child.ID != "urn:record-items:line" || child.Index == nil || *child.Index != i || len(child.Children) != 2 || child.Children[0].ID != "urn:record-items:quantity" || child.Children[1].ID != "urn:record-items:price" {
			t.Fatal("record item path lost", child)
		}
	}
	if run.ReadObservation == nil || run.ReadObservation.ResolvedReads != 4 {
		t.Fatal("record collection reads lost original element indexes", run.ReadObservation)
	}
	if len(saved.Plan.Records) != 2 {
		t.Fatal("transitive record declaration lost", saved.Plan.Records)
	}
}

func TestRecordCollectionAcrossBindRetainsDecodedOrigins(t *testing.T) {
	source := append(nestedLineItemsSource(t), []byte("activity Copy(Basket) -> Basket id \"urn:record-items:copy\" computes `return input`\nbind Copy.result -> Total.input\n")...)
	raw := json.RawMessage(`{"items":[{"quantity":2,"price":7},{"quantity":2,"price":7}]}`)
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1", Cases: []CompositionCase{{Inputs: map[string]json.RawMessage{"Copy": raw}, Expected: map[string]json.RawMessage{"Copy": raw, "Total": json.RawMessage(`28`)}}}}
	prior, err := GenerateComposition(context.Background(), "bind.gooo", source, suite, "")
	if err != nil {
		t.Fatal(err)
	}
	run, err := ExecuteComposition(context.Background(), "bind.gooo", source, prior, suite, nativeTool())
	if err != nil || !run.RuntimeReplayed || run.FinitePassed != 2 || run.FiniteTotal != 2 || run.ModelCalls != 0 {
		t.Fatalf("nested bind failed: %v stage=%s build=%s", err, run.Stage, run.Build.Diagnostics)
	}
	if len(run.Traces) != 1 || len(run.Traces[0].Deliveries) != 2 || len(run.Traces[0].Deliveries[0].ActualFields[0].Children) != 2 {
		t.Fatal("nested bind values/source paths lost")
	}
	if run.ReadObservation == nil || run.ReadObservation.ResolvedReads != 2 || run.ReadObservation.UnresolvedReads != 0 {
		t.Fatal("bind lost original backing element identity", run.ReadObservation)
	}
}

func TestOptionalRecordValuesPreserveAbsenceAndOrder(t *testing.T) {
	source := []byte("package optional\nnamespace optional\nentity Line id \"urn:optional:line\" fields {\nfield n id \"urn:optional:n\" type integer required one\n}\n" +
		"entity Envelope id \"urn:optional:envelope\" fields {\nfield primary id \"urn:optional:primary\" type Line optional one\nfield lines id \"urn:optional:lines\" type Line optional many\n}\n" +
		"activity Copy(Envelope) -> Envelope id \"urn:optional:copy\" computes `return input`\n")
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1"}
	for _, raw := range []string{`{}`, `{"primary":{"n":9007199254740993},"lines":[]}`, `{"lines":[{"n":3},{"n":3}]}`} {
		suite.Cases = append(suite.Cases, CompositionCase{Inputs: map[string]json.RawMessage{"Copy": json.RawMessage(raw)}, Expected: map[string]json.RawMessage{"Copy": json.RawMessage(raw)}})
	}
	prior, err := GenerateComposition(context.Background(), "optional.gooo", source, suite, "")
	if err != nil {
		t.Fatal(err)
	}
	run, err := ExecuteComposition(context.Background(), "optional.gooo", source, prior, suite, nativeTool())
	if err != nil || !run.RuntimeReplayed || run.FinitePassed != 3 || run.FiniteTotal != 3 {
		t.Fatalf("optional record copy: %v stage=%s build=%s", err, run.Stage, run.Build.Diagnostics)
	}
	if field := run.Traces[0].Deliveries[0].ActualFields[0]; field.Present == nil || *field.Present {
		t.Fatal("absent optional record became present", field)
	}
	if field := run.Traces[1].Deliveries[0].ActualFields[1]; field.Present == nil || !*field.Present || string(field.Value) != "[]" {
		t.Fatal("present empty record collection became absent", field)
	}
}
