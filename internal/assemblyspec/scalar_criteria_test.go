package assemblyspec

import (
	"reflect"
	"testing"
)

func TestScalarCriteriaCanonicalInputsAndCrossPhaseLimits(t *testing.T) {
	s := Spec{Choices: []Choice{{ID: "direction", Kind: "operand_order", Intent: "Subtract 11"}}, MaxAttempts: 2,
		Cases: []Case{{Input: 12, Expected: 1}}, HoldoutCases: []Case{{Input: 21, Expected: 10}},
		Criteria:    &Criteria{Predicates: []Criterion{{ID: "difference", Activity: "Difference"}}, Inputs: []string{"[11]"}, HoldoutInputs: []string{"[20]"}},
		Observation: &Observation{CriteriaInputs: []string{"[17]"}, CriteriaStrategy: "distinguish_then_check", MaxCandidates: 2, MaxRounds: 1}}
	raw, err := s.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	back, err := DecodeCanonical(raw)
	if err != nil || !reflect.DeepEqual(back, &s) || !IsScalarCriteria(back) {
		t.Fatal(err, back)
	}
	for _, change := range []func(*Spec){
		func(s *Spec) { s.Criteria.HoldoutInputs = []string{"[12]"} },
		func(s *Spec) { s.HoldoutCases[0].Input = 11 },
		func(s *Spec) { s.Observation.CriteriaInputs = []string{"[21]"} },
		func(s *Spec) { s.ValueCases = []ValueCase{{Inputs: "[1]", Expected: "{}"}} },
		func(s *Spec) { s.Seed = "unowned-order" },
		func(s *Spec) { s.ConditionCases = []ConditionCase{{ChoiceID: "direction", Input: 0, Expected: true}} },
		func(s *Spec) {
			s.Choices = append(s.Choices, Choice{ID: "field", Kind: "field_value", Alternative: "0", Intent: "mixed"})
		},
	} {
		copy := s.Clone()
		change(copy)
		if copy.Validate() == nil {
			t.Fatal("invalid scalar source contract accepted", copy)
		}
	}
}
