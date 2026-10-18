package bodyexecution

import (
	"context"
	"encoding/json"
	"testing"
)

func TestCollectionMapNativeFaultPrefix(t *testing.T) {
	const source = `package mapping
namespace map_fault
entity Integer id "urn:gooo:type:integer"
entity Batch id "urn:map-fault:batch" fields {field values id "urn:map-fault:values" type integer required many}
activity Invert(Integer) -> Integer id "urn:map-fault:invert" computes "return 100/input"
activity Build(Batch) -> Batch id "urn:map-fault:build" computes "return Batch{values:map_values(input.values,Invert)}"
activity Read(Batch) -> Integer id "urn:map-fault:read" computes "return input.values[0]"
bind Build.result -> Read.input
`
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1", Cases: []CompositionCase{
		{Inputs: map[string]json.RawMessage{"Build": json.RawMessage(`{"values":[2,0,4]}`)}, Expected: map[string]json.RawMessage{"Read": json.RawMessage(`0`)}},
		{Inputs: map[string]json.RawMessage{"Build": json.RawMessage(`{"values":[4,2]}`)}, Expected: map[string]json.RawMessage{"Read": json.RawMessage(`25`)}},
	}}
	p, err := GenerateCompositionWithOptions(context.Background(), "fault.gooo", []byte(source), suite, CompositionOptions{EntryActivity: "Read"})
	if err != nil {
		t.Fatal(err)
	}
	r, err := ExecuteComposition(context.Background(), "fault.gooo", []byte(source), p, suite, "")
	if err != nil || r.FinitePassed != 1 || !r.RuntimeReplayed {
		t.Fatal(err, r.Failure)
	}
	if r.ReadObservation.InterruptedConstructions != 1 || len(r.Traces[0].Constructions) != 1 {
		t.Fatal(r.ReadObservation)
	}
	created := r.Traces[0].Constructions[0]
	if created.Completed || created.ElementCount == nil || *created.ElementCount != 3 || len(created.Elements) != 2 || !created.Elements[0].Completed || created.Elements[1].Completed || len(r.Traces[0].Reads) != 2 {
		t.Fatal("attempted prefix", created)
	}
	if err := validateRecordedNativeOrigins(r); err != nil {
		t.Fatal("partial origin validation", err)
	}
}
