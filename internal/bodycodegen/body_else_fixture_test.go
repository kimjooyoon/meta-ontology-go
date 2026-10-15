package bodycodegen

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestElseControlFixtureNamesTheCurrentProgramGraph(t *testing.T) {
	source, err := os.ReadFile("../../examples/else-layout/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	report, err := InspectTypedValueOrigins("main.gooo", source, "Adjust")
	if err != nil {
		t.Fatal(err)
	}
	current, err := TypedSourceControlInputs(report, "source/control:Inspect")
	if err != nil {
		t.Fatal(err)
	}
	var envelope struct {
		Inputs []map[string]SourceControlColumns `json:"inputs"`
	}
	if err := json.Unmarshal(current, &envelope); err != nil || len(envelope.Inputs) != 1 {
		t.Fatal("current source graph", err)
	}
	raw, err := os.ReadFile("../../examples/else-layout/source-control-case.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Inputs map[string]SourceControlColumns `json:"inputs"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil || len(fixture.Cases) != 1 ||
		!reflect.DeepEqual(envelope.Inputs[0], fixture.Cases[0].Inputs) {
		t.Fatal("reader fixture inherited another program's source or graph", err)
	}
}
