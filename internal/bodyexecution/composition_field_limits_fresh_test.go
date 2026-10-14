package bodyexecution

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

func TestGoooFieldLimits13601(t *testing.T) {
	const source = `package fieldlimits13601
namespace fieldlimits13601
entity Triple id "urn:fieldlimits13601:triple" fields {
field value id "urn:fieldlimits13601:value" type integer required one
field flag id "urn:fieldlimits13601:flag" type boolean required one
field label id "urn:fieldlimits13601:label" type string required one
}
activity Tick(Triple)->Triple id "urn:fieldlimits13601:tick" computes ` + "`let next=input;next.value=next.value+2111;return next`" + `
activity Build(Triple)->Triple id "urn:fieldlimits13601:build" computes ` + "`let current=input;current.flag=!current.flag;current.label=current.label+\"!\";let i=0;for i=0;i<1024;i=i+1 {current=Tick(current);current=Tick(current);current=Tick(current)};return current`" + `
`
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1"}
	for i := range 24 {
		input, _ := json.Marshal(map[string]any{"value": 2203 + i, "flag": i%2 == 0, "label": fmt.Sprintf("sample%d", i)})
		expected, _ := json.Marshal(map[string]any{"value": 2203 + i + 3072*2111, "flag": i%2 != 0, "label": fmt.Sprintf("sample%d!", i)})
		suite.Cases = append(suite.Cases, CompositionCase{Inputs: map[string]json.RawMessage{"Build": input}, Expected: map[string]json.RawMessage{"Build": expected}})
	}
	o := map[string]any{"source": source, "cases": suite}
	defer saveReadObservation(t, "field-limits13601.json", o)
	program, err := GenerateCompositionWithOptions(context.Background(), "fieldlimits13601.gooo", []byte(source), suite, CompositionOptions{EntryActivity: "Build"})
	o["composition"] = program
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := ExecuteComposition(context.Background(), "fieldlimits13601.gooo", []byte(source), program, suite, "")
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

func TestGoooFieldFault13627(t *testing.T) {
	const source = `package fieldfault13627
namespace fieldfault13627
entity Integer id "urn:gooo:type:integer"
entity Input id "urn:fieldfault13627:input" fields {
field value id "urn:fieldfault13627:value" type integer required one
field fail id "urn:fieldfault13627:fail" type boolean required one
}
activity Build(Input)->Integer id "urn:fieldfault13627:build" computes ` + "`let current=input;if input.fail {current.value=current.value/(input.value-input.value)};return current.value`" + `
`
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1", Cases: []CompositionCase{
		{Inputs: map[string]json.RawMessage{"Build": json.RawMessage(`{"value":2237,"fail":false}`)}, Expected: map[string]json.RawMessage{"Build": json.RawMessage(`2237`)}},
		{Inputs: map[string]json.RawMessage{"Build": json.RawMessage(`{"value":2239,"fail":true}`)}, Expected: map[string]json.RawMessage{"Build": json.RawMessage(`0`)}},
	}}
	o := map[string]any{"source": source, "cases": suite}
	defer saveReadObservation(t, "field-fault13627.json", o)
	program, err := GenerateCompositionWithOptions(context.Background(), "fieldfault13627.gooo", []byte(source), suite, CompositionOptions{EntryActivity: "Build"})
	if err != nil {
		t.Fatal(err)
	}
	o["composition"] = program
	runtime, err := ExecuteComposition(context.Background(), "fieldfault13627.gooo", []byte(source), program, suite, "")
	o["runtime"] = runtime
	if err != nil {
		t.Fatal(err)
	}
	if runtime.FinitePassed != 1 || !hasCompositionFault(runtime) || runtime.FieldObservation == nil || runtime.FieldObservation.Interrupted != 1 || runtime.FieldObservation.Status != "PARTIAL" || !runtime.RuntimeReplayed {
		t.Fatal("actual fault must retain interrupted field write", runtime.FinitePassed, runtime.FieldObservation)
	}
	if err := validateRecordedFieldObservations(runtime); err != nil {
		t.Fatal(err)
	}
	t.Log("fresh2 inputs:one actual result andone intended arithmetic fault;interrupted field write retains prior completed assignment;2-run equality;models0")
}
