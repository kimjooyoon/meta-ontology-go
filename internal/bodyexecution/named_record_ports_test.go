package bodyexecution

import (
	"context"
	"encoding/json"
	"testing"
)

func TestNamedRecordPortsRetainPositions(t *testing.T) {
	source := `package namedports
namespace namedports
entity Integer id "urn:gooo:type:integer"
entity Pair id "urn:namedports:pair" fields {
field value id "urn:namedports:value" type integer required one
}
activity Carry(쌍 Pair) -> Pair id "urn:namedports:carry" computes "return 쌍"
activity Combine(first Pair, delta Integer, 마지막 Pair) -> Pair id "urn:namedports:combine" computes ` + "`" + `
let result = Carry(first)
result.value += 마지막.value + delta
return result
` + "`\n"
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1", Cases: []CompositionCase{{
		Inputs:   map[string]json.RawMessage{"Combine.input0": json.RawMessage(`{"value":5}`), "Combine.input1": json.RawMessage(`2`), "Combine.input2": json.RawMessage(`{"value":7}`)},
		Expected: map[string]json.RawMessage{"Combine": json.RawMessage(`{"value":14}`)},
	}}}
	program, err := GenerateCompositionWithOptions(context.Background(), "named-ports.gooo", []byte(source), suite, CompositionOptions{EntryActivity: "Combine"})
	if err != nil {
		t.Fatal(err)
	}
	inventory, err := compositionFieldBindings(program)
	if err != nil {
		t.Fatal(err)
	}
	wanted := map[string]string{"first": "input0", "마지막": "input2", "쌍": "input"}
	for _, binding := range inventory.bindings {
		if binding.VariableOrdinal >= 0 {
			continue
		}
		if wanted[binding.Name] != binding.InputPort {
			t.Fatal("authored name changed wire identity", binding)
		}
		delete(wanted, binding.Name)
		if err := validateFieldBinding(binding); err != nil {
			t.Fatal(err)
		}
		binding.InputPort = binding.Name
		if err := validateFieldBinding(binding); err == nil {
			t.Fatal("display name accepted as canonical port")
		}
	}
	if len(wanted) != 0 {
		t.Fatal("missing root/helper parameter", wanted)
	}
}
