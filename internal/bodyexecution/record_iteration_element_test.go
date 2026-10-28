package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"slices"
	"strings"
	"testing"
)

func recordIterationRuntime(t *testing.T, source []byte, cases []byte) CompositionRuntime {
	t.Helper()
	suite, err := DecodeCompositionCases(cases)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := GenerateCompositionWithOptions(context.Background(), "elements.gooo", source, suite, CompositionOptions{EntryActivity: "Total"})
	if err != nil {
		t.Fatal(err)
	}
	r, err := ExecuteComposition(context.Background(), "elements.gooo", source, plan, suite, "")
	if err != nil {
		t.Fatal(err)
	}
	if !r.RuntimeReplayed || r.ModelCalls != 0 || r.FinitePassed != len(suite.Cases) {
		t.Fatal("independent expected values, replay or zero inference changed", r.FinitePassed, r.FiniteTotal)
	}
	return r
}

func recordIterationExample(t *testing.T) ([]byte, []byte) {
	t.Helper()
	source, err := os.ReadFile("../../examples/record-collections/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	cases, err := os.ReadFile("../../examples/record-collections/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	return source, cases
}

func TestRecordIterationElementLinks(t *testing.T) {
	source, cases := recordIterationExample(t)
	r := recordIterationRuntime(t, source, cases)
	want := [][]int{nil, {0, 0}, {0, 0, 1, 1}, {0, 0}}
	if r.BridgeObservation.Schema != "gooo/native-field-local-bridge-summary/v2" ||
		r.BridgeObservation.ElementLinks != 8 || r.BridgeObservation.MissingElements != 0 ||
		r.BridgeObservation.Status != "COMPLETE" || r.ReadObservation.RecordedReads != 4 {
		t.Fatal("actual element-link population differs", r.BridgeObservation, r.ReadObservation)
	}
	for c, trace := range r.Traces {
		if !slices.Equal(trace.Bridge.ElementReads, want[c]) {
			t.Fatal("field-to-read references are not original case-local occurrences", c, trace.Bridge.ElementReads)
		}
	}
	for _, kind := range []string{"field-writes", "field-uses"} {
		rows, report, err := SelectRecordedNativeOrigins(r, NativeOriginQuery{Kind: kind, Limit: 128})
		if err != nil || len(rows) != 8 || report.Schema != "gooo/native-origin-input/v2" || report.ProducerExecutions != 0 || report.ProducerModelCalls != 0 {
			t.Fatal("saved row transport lost scope or reran the producer", err, report)
		}
		for i, raw := range rows {
			var fact CompositionFieldFact
			if err := json.Unmarshal(raw, &fact); err != nil {
				t.Fatal(err)
			}
			index := []int{0, 0, 0, 0, 1, 1, 0, 0}[i]
			if fact.Element == nil || fact.ElementReadIndex == nil || *fact.ElementReadIndex != index ||
				fact.Element.Status != "RESOLVED" || fact.Element.OriginIndex != index ||
				fact.Element.FieldID != "urn:record-items:items" || fact.ElementReason != "" {
				t.Fatal("same-valued inputs lost their original position or collection field", fact)
			}
		}
		if kind == "field-uses" {
			checkIterationElementConsumer(t, rows)
		}
	}
	q := NativeOriginQuery{Kind: "field-uses", Limit: 1}
	if _, _, err := SelectRecordedFieldLineage(r, q); err == nil {
		t.Fatal("v1 field columns silently discarded element references")
	}
	if _, _, err := SelectRecordedBridgeLineage(r, q); err == nil {
		t.Fatal("v1 bridge columns silently discarded element references")
	}
	checkIterationElementContradictions(t, r)
	t.Log("4 original cases pass, including empty, equal-valued positions0/1 and >2^53;8 field assignments reference4 existing reads;2-run replay;old columns decline richer facts;producer executions0/model0 during query")
}

func checkIterationElementConsumer(t *testing.T, rows []json.RawMessage) {
	t.Helper()
	source, err := os.ReadFile("testdata/record-element-locate.gooo")
	if err != nil {
		t.Fatal(err)
	}
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1"}
	for i, row := range rows {
		var fact CompositionFieldFact
		if err := json.Unmarshal(row, &fact); err != nil {
			t.Fatal(err)
		}
		input, _ := json.Marshal(map[string]any{"position": fact.Element.OriginIndex,
			"container": fact.Element.FieldID, "status": fact.Element.Status})
		expected, _ := json.Marshal([]int{0, 0, 0, 0, 1, 1, 0, 0}[i])
		suite.Cases = append(suite.Cases, CompositionCase{Inputs: map[string]json.RawMessage{"Locate": input},
			Expected: map[string]json.RawMessage{"Locate": expected}})
	}
	// These two controls are authored refusals, separate from the eight real rows.
	for _, input := range []string{
		`{"position":0,"container":"urn:record-items:items","status":"UNRESOLVED"}`,
		`{"position":0,"container":"urn:other:collection","status":"RESOLVED"}`,
	} {
		suite.Cases = append(suite.Cases, CompositionCase{Inputs: map[string]json.RawMessage{"Locate": json.RawMessage(input)},
			Expected: map[string]json.RawMessage{"Locate": json.RawMessage(`-1`)}})
	}
	plan, err := GenerateCompositionWithOptions(context.Background(), "locate.gooo", source, suite, CompositionOptions{EntryActivity: "Locate"})
	if err != nil {
		t.Fatal(err)
	}
	r, err := ExecuteComposition(context.Background(), "locate.gooo", source, plan, suite, "")
	if err != nil || r.FinitePassed != 10 || !r.RuntimeReplayed || r.ModelCalls != 0 {
		t.Fatal("Gooo did not own location selection and unknown refusal", err, r.FinitePassed)
	}
	t.Log("source-owned consumer:8 real location rows +2 authored refusal controls;all10 native decisions match independent expectations;producer not replayed during query;model0")
}

func checkIterationElementContradictions(t *testing.T, r CompositionRuntime) {
	t.Helper()
	changes := []func(*CompositionRuntime){
		func(r *CompositionRuntime) { r.Traces[2].Bridge.ElementReads = r.Traces[2].Bridge.ElementReads[:3] },
		func(r *CompositionRuntime) { r.Traces[2].Bridge.ElementReads[0] = -1 },
		func(r *CompositionRuntime) { r.Traces[2].Bridge.ElementReads[0] = -2 },
		func(r *CompositionRuntime) { r.Traces[2].Bridge.ElementReads[0] = 2 },
		func(r *CompositionRuntime) { r.Traces[2].Bridge.ElementReads[1] = 1 },
		func(r *CompositionRuntime) {
			r.Traces[2].Bridge.ElementReads[2] = 0
			r.Traces[2].Bridge.ElementReads[3] = 0
		},
		func(r *CompositionRuntime) { r.ReadSites[0].Kind = "index_value" },
		func(r *CompositionRuntime) { r.ReadSites[0].ElementType = "wrong-record" },
		func(r *CompositionRuntime) { r.ReadSites[0].Source.Start++ },
	}
	raw, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	for i, change := range changes {
		var altered CompositionRuntime
		if err := json.Unmarshal(raw, &altered); err != nil {
			t.Fatal(err)
		}
		change(&altered)
		if err := validateBridgeElementReads(altered.Traces[2], altered); err == nil {
			t.Fatal("contradictory element reference accepted", i)
		}
	}
}

func TestRecordIterationNestedElementLinks(t *testing.T) {
	source, _ := recordIterationExample(t)
	source = []byte(strings.Replace(string(source), "field items", "field others id \"urn:record-items:others\" type Line required many\n field items", 1))
	source = []byte(strings.Replace(string(source), "total = total + item.quantity * item.price", "let copied = item; for extra in input.others { total = total + copied.quantity * extra.price }", 1))
	cases := []byte(`{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"Total":{"items":[{"quantity":2,"price":7},{"quantity":2,"price":7}],"others":[{"quantity":2,"price":7},{"quantity":2,"price":7}]}},"expected":{"Total":56}}]}`)
	r := recordIterationRuntime(t, source, cases)
	want := []int{0, 0, -1, -1, 1, 1, 2, 2, 3, 3, -1, -1, 4, 4, 5, 5}
	if !slices.Equal(r.Traces[0].Bridge.ElementReads, want) || r.BridgeObservation.ElementLinks != 12 {
		t.Fatal("nested or copied records guessed an element parent", r.Traces[0].Bridge.ElementReads, r.BridgeObservation)
	}
	rows, _, err := SelectRecordedNativeOrigins(r, NativeOriginQuery{Kind: "field-writes", Limit: 128})
	if err != nil || len(rows) != len(want) {
		t.Fatal("nested field facts", err, len(rows))
	}
	for _, raw := range rows {
		var fact CompositionFieldFact
		if err := json.Unmarshal(raw, &fact); err != nil {
			t.Fatal(err)
		}
		switch fact.VariableName {
		case "copied":
			if fact.Element != nil || fact.ParentAssignment < 0 {
				t.Fatal("copy should retain its original field parent", fact)
			}
		case "item", "extra":
			field := map[string]string{"item": "urn:record-items:items", "extra": "urn:record-items:others"}[fact.VariableName]
			if fact.Element == nil || fact.Element.FieldID != field {
				t.Fatal("equal-valued collections were confused", fact)
			}
		default:
			t.Fatal("unexpected lexical binding", fact)
		}
	}
}

func TestRecordIterationElementReadLimit(t *testing.T) {
	source, _ := recordIterationExample(t)
	source = []byte(strings.Replace(string(source), "field items", "field values id \"urn:record-items:values\" type integer required many\n field items", 1))
	prefix := "let ignored=0; let turn=0; for turn=0;turn<1024;turn=turn+1 {" +
		strings.Repeat("ignored=input.values[0];", 64) + "}; for item in"
	source = []byte(strings.Replace(string(source), "for item in", prefix, 1))
	values := []int{0}
	input, err := json.Marshal(map[string]any{"schema": "gooo/body-composition-cases/v1", "cases": []any{
		map[string]any{"inputs": map[string]any{"Total": map[string]any{"items": []any{map[string]int{"quantity": 2, "price": 7}}, "values": values}}, "expected": map[string]int{"Total": 14}},
	}})
	if err != nil {
		t.Fatal(err)
	}
	r := recordIterationRuntime(t, source, input)
	if !slices.Equal(r.Traces[0].Bridge.ElementReads, []int{-2, -2}) ||
		r.ReadObservation.RecordedReads != 65536 || r.ReadObservation.DroppedReads != 1 ||
		r.BridgeObservation.Status != "PARTIAL" || r.BridgeObservation.ElementLinks != 0 ||
		r.BridgeObservation.MissingElements != 2 {
		t.Fatal("dropped range read was replaced by an older recorded row", r.Traces[0].Bridge.ElementReads, r.BridgeObservation, r.ReadObservation)
	}
	rows, _, err := SelectRecordedNativeOrigins(r, NativeOriginQuery{Kind: "field-writes", Limit: 128})
	if err != nil || len(rows) != 2 {
		t.Fatal("partial element facts", err)
	}
	for _, raw := range rows {
		var fact CompositionFieldFact
		if err := json.Unmarshal(raw, &fact); err != nil || fact.Element != nil || fact.ElementReadIndex != nil || fact.ElementReason != "READ_LIMIT" {
			t.Fatal("missing observation fabricated an element origin", err, fact)
		}
	}
}
