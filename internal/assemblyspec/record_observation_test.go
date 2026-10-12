package assemblyspec

import "testing"

func TestRecordObservationContractBoundsAndInputSeparation(t *testing.T) {
	s := Spec{Choices: []Choice{{ID: "value", Kind: "field_value", Alternative: "input", Intent: "Keep the input."}},
		ValueCases:        []ValueCase{{Inputs: "[1]", Expected: `{"value":1}`}},
		ValueHoldoutCases: []ValueCase{{Inputs: "[0]", Expected: `{"value":0}`}}, MaxAttempts: 2,
		Observation: &Observation{ValueInputs: []string{"[9007199254740993]"}, MaxCandidates: 2, MaxRounds: 1}}
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
	changes := []func(*Spec){
		func(x *Spec) { x.Observation.ValueInputs = []string{"[-0]"} },
		func(x *Spec) { x.Observation.ValueInputs = []string{"[2]", "[2]"} },
		func(x *Spec) { x.Observation.ValueInputs = []string{"[ 2 ]"} },
		func(x *Spec) { x.Observation.ValueInputs = nil },
		func(x *Spec) { x.Observation.Inputs = []int64{3} },
		func(x *Spec) { x.Observation.MaxCandidates = 65 },
		func(x *Spec) { x.Observation.MaxRounds = 9 },
	}
	for i, change := range changes {
		x := s.Clone()
		change(x)
		if err := x.Validate(); err == nil {
			t.Fatalf("invalid record observation %d accepted", i)
		}
	}
	x := s.Clone()
	x.Observation.ValueInputs[0] = "[2]"
	if s.Observation.ValueInputs[0] != "[9007199254740993]" {
		t.Fatal("clone shared record questions")
	}
}
