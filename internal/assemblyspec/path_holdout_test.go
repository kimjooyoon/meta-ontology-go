package assemblyspec

import "testing"

func TestPathHoldoutBoundsAndDisjointSelectionInputs(t *testing.T) {
	s := Spec{Choices: []Choice{{ID: "direction", Kind: "operand_order", Intent: "Choose the direction."}},
		Cases: []Case{{Input: 19, Expected: 0}}, HoldoutCases: []Case{{Input: 9007199254740993, Expected: 9007199254740974}}, MaxAttempts: 2}
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
	clone := s.Clone()
	clone.HoldoutCases[0].Expected++
	if s.HoldoutCases[0].Expected != 9007199254740974 {
		t.Fatal("holdout storage was shared or rounded")
	}
	for _, change := range []func(*Spec){
		func(s *Spec) { s.Cases = nil },
		func(s *Spec) { s.HoldoutCases[0].Input = 19 },
		func(s *Spec) { s.HoldoutCases = append(s.HoldoutCases, s.HoldoutCases[0]) },
		func(s *Spec) { s.ConditionCases = []ConditionCase{{ChoiceID: "direction", Input: 9007199254740993}} },
		func(s *Spec) {
			s.Observation = &Observation{Inputs: []int64{9007199254740993}, MaxCandidates: 2, MaxRounds: 1}
		},
	} {
		bad := s.Clone()
		change(bad)
		if err := bad.Validate(); err == nil {
			t.Fatal("invalid holdout separation accepted", bad)
		}
	}
	s.HoldoutCases = nil
	for i := int64(100); i < 227; i++ {
		s.HoldoutCases = append(s.HoldoutCases, Case{Input: i})
	}
	if err := s.Validate(); err != nil {
		t.Fatal("128 total source cases rejected", err)
	}
	s.HoldoutCases = append(s.HoldoutCases, Case{Input: 227})
	if err := s.Validate(); err == nil {
		t.Fatal("source case bound exceeded")
	}
}
