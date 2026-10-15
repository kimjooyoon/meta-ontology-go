package bodyexecution

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

func TestGoooFieldLimits13811(t *testing.T) {
	const source = `package fieldlimits13811
namespace fieldlimits13811
entity State id "urn:fieldlimits13811:triple" fields {
field points id "urn:fieldlimits13811:points" type integer required one
field gate id "urn:fieldlimits13811:gate" type boolean required one
field tag id "urn:fieldlimits13811:tag" type string required one
}
activity Step(State)->State id "urn:fieldlimits13811:tick" computes ` + "`let updated=input;updated.points=updated.points+2207;return updated`" + `
activity Build(State)->State id "urn:fieldlimits13811:build" computes ` + "`let state=input;state.gate=!state.gate;state.tag=state.tag+\"?\";let i=0;for i=0;i<1024;i=i+1 {state=Step(state);state=Step(state);state=Step(state)};return state`" + `
`
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1"}
	for i := range 24 {
		input, _ := json.Marshal(map[string]any{"points": 2411 + i, "gate": i%2 == 0, "tag": fmt.Sprintf("batch%d", i)})
		expected, _ := json.Marshal(map[string]any{"points": 2411 + i + 3072*2207, "gate": i%2 != 0, "tag": fmt.Sprintf("batch%d?", i)})
		suite.Cases = append(suite.Cases, CompositionCase{Inputs: map[string]json.RawMessage{"Build": input}, Expected: map[string]json.RawMessage{"Build": expected}})
	}
	o := map[string]any{"source": source, "cases": suite}
	defer saveReadObservation(t, "field-limits13811.json", o)
	program, err := GenerateCompositionWithOptions(context.Background(), "fieldlimits13811.gooo", []byte(source), suite, CompositionOptions{EntryActivity: "Build"})
	o["composition"] = program
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := ExecuteComposition(context.Background(), "fieldlimits13811.gooo", []byte(source), program, suite, "")
	o["runtime"] = runtime
	if err != nil {
		t.Fatal(err)
	}
	summary := runtime.FieldObservation
	if runtime.FinitePassed != 24 || runtime.ModelCalls != 0 || !runtime.RuntimeReplayed || summary == nil || summary.Status != "PARTIAL" || summary.RecordedWrites != 4096 || summary.RecordedUses != 65536 || summary.Frames != 65536 || summary.WritesDropped < 1 || summary.UsesDropped < 1 || summary.FramesDropped < 1 {
		t.Fatal("native field caps and actual finite outcomes", runtime.FinitePassed, summary)
	}
	if err := validateRecordedFieldObservations(runtime); err != nil {
		t.Fatal(err)
	}
	t.Log("24 fresh Gooo inputs,73728 fixed helper calls;actual integer/boolean/string outputs24/24;4096writes/65536uses/65536frames bounded with explicit drops;2-run replay;models0")
}
