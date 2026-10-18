package bodyexecution

import (
	"context"
	"encoding/json"
	"testing"
)

func TestCollectionFilterNativeFaultPrefix(t *testing.T) {
	const source = `package filtering
namespace filter_fault
entity Integer id "urn:gooo:type:integer"
entity Boolean id "urn:gooo:type:boolean"
entity Batch id "urn:filter-fault:batch" fields {field values id "urn:filter-fault:values" type integer required many}
activity Keep(Integer) -> Boolean id "urn:filter-fault:keep" computes "return input/(input-1)>0"
activity Build(Batch) -> Batch id "urn:filter-fault:build" computes "return Batch{values:filter_values(input.values,Keep)}"
activity Read(Batch) -> Integer id "urn:filter-fault:read" computes "return input.values[0]"
bind Build.result -> Read.input
`
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1", Cases: []CompositionCase{
		{Inputs: map[string]json.RawMessage{"Build": json.RawMessage(`{"values":[1]}`)}, Expected: map[string]json.RawMessage{"Read": json.RawMessage(`0`)}},
		{Inputs: map[string]json.RawMessage{"Build": json.RawMessage(`{"values":[2,1]}`)}, Expected: map[string]json.RawMessage{"Read": json.RawMessage(`0`)}},
		{Inputs: map[string]json.RawMessage{"Build": json.RawMessage(`{"values":[2,3]}`)}, Expected: map[string]json.RawMessage{"Read": json.RawMessage(`2`)}},
	}}
	program, err := GenerateCompositionWithOptions(context.Background(), "filter-fault.gooo", []byte(source), suite, CompositionOptions{EntryActivity: "Read"})
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := ExecuteComposition(context.Background(), "filter-fault.gooo", []byte(source), program, suite, "")
	if err != nil || runtime.FinitePassed != 1 || !runtime.RuntimeReplayed {
		t.Fatal(err, runtime.Failure, runtime.FinitePassed)
	}
	if runtime.ReadObservation == nil || runtime.ReadObservation.InterruptedConstructions != 2 {
		t.Fatal("fault observations", runtime.ReadObservation)
	}
	first, second := runtime.Traces[0].Constructions[0], runtime.Traces[1].Constructions[0]
	if first.Completed || len(first.Elements) != 0 || first.ReadEnd != 1 || second.Completed || len(second.Elements) != 1 || second.Elements[0].InputIndex == nil || *second.Elements[0].InputIndex != 0 || second.ReadEnd != 2 {
		t.Fatal("retained prefix before predicate fault", first, second)
	}
	if err := validateRecordedNativeOrigins(runtime); err != nil {
		t.Fatal("partial origin validation", err)
	}
}
