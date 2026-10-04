package bodyexecution

import (
	"context"
	"encoding/json"
	"testing"
)

func TestCompositionNativeExternalRecordCopyAndLocalReplacement(t *testing.T) {
	source := []byte("package records\nnamespace recordinput\n" +
		"entity Text id \"recordinput://text\"\n" +
		"entity Item id \"recordinput://item\" fields {\n" +
		"field name id \"recordinput://name\" type string required one\n" +
		"field state id \"recordinput://state\" type string required one\n}\n" +
		"activity Start(Item) -> Item computes `let copy = input\n" +
		"copy = Item{name: copy.name, state: copy.state + \"!\"}\nreturn copy`\n" +
		"activity End(Item) -> Text computes `return input.name + \":\" + input.state`\n" +
		"bind Start.result -> End.input\n")
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1", Cases: []CompositionCase{{
		Inputs: map[string]json.RawMessage{"Start": json.RawMessage(`{"state":"ready","name":"한글"}`)},
		Expected: map[string]json.RawMessage{"Start": json.RawMessage(`{"name":"한글","state":"ready!"}`),
			"End": json.RawMessage(`"한글:ready!"`)}}}}
	prior, err := GenerateComposition(context.Background(), "input.gooo", source, suite, "")
	if err != nil {
		t.Fatal(err)
	}
	run, err := ExecuteComposition(context.Background(), "input.gooo", source, prior, suite, nativeTool())
	if err != nil || run.FinitePassed != 2 || run.FiniteTotal != 2 || !run.RuntimeReplayed {
		t.Fatalf("external record: %v %+v", err, run)
	}
	first := run.Traces[0].Deliveries[0]
	if len(first.InputFields) != 2 || first.InputFields[0].ID != "recordinput://name" ||
		string(first.InputFields[1].Value) != `"ready"` || string(first.ActualFields[1].Value) != `"ready!"` {
		t.Fatal("record input or local copy observation differs", first)
	}
	for _, raw := range []string{`{"name":"x"}`, `{"name":"x","state":0}`} {
		suite.Cases[0].Inputs["Start"] = json.RawMessage(raw)
		result, err := GenerateComposition(context.Background(), "bad-input.gooo", source, suite, "missing-model.json")
		if err == nil || len(result.Steps) > 0 || result.Model != nil {
			t.Fatal("invalid record input reached generation/model", err)
		}
	}
}
