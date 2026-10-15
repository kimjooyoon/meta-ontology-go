package bodyexecution

import (
	"context"
	"encoding/json"
	"testing"
)

func TestReadIndexOrder9361(t *testing.T) {
	const source = `package indexed
namespace indexed9361
entity Integer id "urn:gooo:type:integer"
entity Input id "urn:indexed9361:input" fields {
 field values id "urn:indexed9361:values" type integer required many
 field order id "urn:indexed9361:order" type integer required many
 field position id "urn:indexed9361:position" type integer required one
}
entity Output id "urn:indexed9361:output" fields {
 field item id "urn:indexed9361:item" type integer optional one
}
activity Seed(Input) -> Input id "urn:indexed9361:seed" computes "return input"
activity ByValue(Input) -> Integer id "urn:indexed9361:value" computes "return input.values[input.order[input.position]]"
activity ByAddress(Input) -> Output id "urn:indexed9361:address" computes "return Output{item: &input.values[input.order[input.position]]}"
bind Seed.result -> ByValue.input
bind Seed.result -> ByAddress.input
`
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1"}
	for _, probe := range []struct{ input, value, record string }{
		{`{"values":[421,431,431],"order":[2,0],"position":0}`, `431`, `{"item":431}`},
		{`{"values":[433,439,443],"order":[2,1],"position":1}`, `439`, `{"item":439}`},
	} {
		suite.Cases = append(suite.Cases, CompositionCase{Inputs: map[string]json.RawMessage{"Seed": json.RawMessage(probe.input)}, Expected: map[string]json.RawMessage{"ByValue": json.RawMessage(probe.value), "ByAddress": json.RawMessage(probe.record)}})
	}
	program, err := GenerateComposition(context.Background(), "순서9361.gooo", []byte(source), suite, "")
	observed := map[string]any{"source": source, "cases": suite, "composition": program}
	defer saveReadObservation(t, "order9361.json", observed)
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := ExecuteComposition(context.Background(), "순서9361.gooo", []byte(source), program, suite, "")
	observed["runtime"] = runtime
	if err != nil || runtime.FinitePassed != 4 || runtime.ReadObservation == nil || runtime.ReadObservation.ResolvedReads != 6 {
		t.Fatal("nested index/address semantics", err, runtime.ReadObservation)
	}
	for c, trace := range runtime.Traces {
		var reads []CompositionReadObservation
		for _, event := range trace.Reads {
			if runtime.ReadSites[event.SiteIndex].ActivityID == "urn:indexed9361:value" {
				reads = append(reads, event)
			}
		}
		if len(reads) != 2 || runtime.ReadInputs[reads[0].Origin.Input].FieldID != "urn:indexed9361:order" || runtime.ReadInputs[reads[1].Origin.Input].FieldID != "urn:indexed9361:values" || reads[1].Origin.Index != 2-c {
			t.Fatal("index operand must be evaluated before outer value", c, reads)
		}
	}
	for _, site := range runtime.ReadSites {
		if site.ActivityID == "urn:indexed9361:address" && site.Source.Expression != "input.order[input.position]" {
			t.Fatal("address itself is not a value read", site)
		}
	}
	t.Log("order9361: native4/4, input origins6/6; inner index executes first, taking an element address keeps original Go semantics")
}
