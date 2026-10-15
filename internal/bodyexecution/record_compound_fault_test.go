package bodyexecution

import (
	"context"
	"encoding/json"
	"testing"
)

func TestRecordCompoundDivisionFaultOrigin(t *testing.T) {
	const source = `package fielddivision
namespace fielddivision
entity State id "urn:fielddivision:state" fields {
field count id "urn:fielddivision:count" type integer required one
field divisor id "urn:fielddivision:divisor" type integer required one
}
activity Divide(원본 State) -> State id "urn:fielddivision:divide" computes ` + "`" + `
let state = 원본
(state).count /= 원본.divisor
return state
` + "`\n"
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1", Cases: []CompositionCase{
		{Inputs: map[string]json.RawMessage{"Divide": json.RawMessage(`{"count":19,"divisor":3}`)}, Expected: map[string]json.RawMessage{"Divide": json.RawMessage(`{"count":6,"divisor":3}`)}},
		{Inputs: map[string]json.RawMessage{"Divide": json.RawMessage(`{"count":23,"divisor":0}`)}, Expected: map[string]json.RawMessage{"Divide": json.RawMessage(`{"count":23,"divisor":0}`)}},
	}}
	observation := map[string]any{"source": source, "cases": suite}
	defer saveReadObservation(t, "record-compound-fault.json", observation)
	program, err := GenerateCompositionWithOptions(context.Background(), "field-division.gooo", []byte(source), suite, CompositionOptions{EntryActivity: "Divide"})
	if err != nil {
		t.Fatal(err)
	}
	observation["composition"] = program
	r, err := ExecuteComposition(context.Background(), "field-division.gooo", []byte(source), program, suite, "")
	observation["runtime"] = r
	if err != nil {
		observation["error"] = err.Error()
		t.Fatal(err)
	}
	if r.FinitePassed != 1 || r.Outcomes == nil || r.Outcomes.Faulted != 1 || r.ModelCalls != 0 {
		t.Fatal("division outcome", r.Outcomes)
	}
	fault := r.Traces[1].Deliveries[0].Fault
	if fault == nil || fault.Kind != "ZERO_DIVISOR" || fault.Left != 23 || fault.Right != 0 || fault.Site.ActivityID != "urn:fielddivision:divide" || fault.Site.Assignment != "/=" {
		t.Fatal("division fault source", fault)
	}
	assertInterruptedCompoundWrite(t, r)
}

func assertInterruptedCompoundWrite(t *testing.T, r CompositionRuntime) {
	t.Helper()
	count := 0
	for _, write := range r.Traces[1].FieldWrites {
		if r.FieldSites[write.SiteIndex].DependencyMode != "compound_update" {
			continue
		}
		count++
		if write.Completed || write.UseEnd-write.UseStart != 2 || write.Completion != -1 {
			t.Fatal("fault fabricated a completed update", write)
		}
	}
	if count != 1 {
		t.Fatal("interrupted compound write population", count)
	}
}
