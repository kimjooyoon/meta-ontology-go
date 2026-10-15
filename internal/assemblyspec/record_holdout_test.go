package assemblyspec

import "testing"

func TestRecordHoldoutSourceContractBoundsAndSeparation(t *testing.T) {
	s := Spec{Choices: []Choice{{ID: "value", Kind: "field_value", Alternative: "input", Intent: "Use the value."}},
		ValueCases:        []ValueCase{{Inputs: "[0]", Expected: `{"value":0}`}},
		ValueHoldoutCases: []ValueCase{{Inputs: "[9007199254740993]", Expected: `{"value":9007199254740993}`}}, MaxAttempts: 2}
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
	changes := []func(*Spec){
		func(x *Spec) { x.ValueHoldoutCases[0].Inputs = "[0]" },
		func(x *Spec) { x.ValueHoldoutCases = append(x.ValueHoldoutCases, x.ValueHoldoutCases[0]) },
		func(x *Spec) { x.ValueHoldoutCases[0].Inputs = "[ 2 ]" },
		func(x *Spec) { x.ValueHoldoutCases[0].Expected = `{"value": 2}` },
		func(x *Spec) { x.ValueCases = nil },
	}
	for i, change := range changes {
		x := s.Clone()
		change(x)
		if err := x.Validate(); err == nil {
			t.Fatalf("invalid held contract %d accepted", i)
		}
	}
	if s.ValueHoldoutCases[0].Inputs != "[9007199254740993]" {
		t.Fatal("held contract clone was shared")
	}
}

func TestRecordHoldoutInputSeparationUsesExactTypedValues(t *testing.T) {
	cases := []struct {
		name, selected, held string
		overlap              bool
	}{
		{"integer zero", "[0]", "[-0]", true},
		{"nested zero", `[{"amount":0}]`, `[{"amount":-0}]`, true},
		{"text spelling", `["0"]`, `["-0"]`, false},
		{"text and integer", `["0"]`, `[0]`, false},
		{"exact large integers", "[9007199254740992]", "[9007199254740993]", false},
		{"int64 endpoints", "[-9223372036854775808]", "[9223372036854775807]", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := Spec{Choices: []Choice{{ID: "value", Kind: "field_value", Alternative: "input", Intent: "Use the value."}},
				ValueCases:        []ValueCase{{Inputs: c.selected, Expected: `{"value":0}`}},
				ValueHoldoutCases: []ValueCase{{Inputs: c.held, Expected: `{"value":0}`}}, MaxAttempts: 2}
			if err := s.Validate(); (err != nil) != c.overlap {
				t.Fatalf("overlap=%v: %v", c.overlap, err)
			}
			if s.ValueCases[0].Inputs != c.selected || s.ValueHoldoutCases[0].Inputs != c.held {
				t.Fatal("validation rewrote source spelling")
			}
		})
	}
}
