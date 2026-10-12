package assemblyspec

import "testing"

func TestCriteriaQuestionContractCloneAndModes(t *testing.T) {
	s := criteriaSpec()
	s.Observation = &Observation{CriteriaInputs: []string{"[0]", "[-7]"}, MaxCandidates: 2, MaxRounds: 1}
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
	c := s.Clone()
	c.Observation.CriteriaInputs[0] = "[1]"
	if s.Observation.CriteriaInputs[0] != "[0]" {
		t.Fatal("criteria question clone aliases source")
	}
	encoded, _ := s.Canonical()
	decoded, err := DecodeCanonical(encoded)
	if err != nil || len(decoded.Observation.CriteriaInputs) != 2 {
		t.Fatal(decoded, err)
	}
	for _, mutate := range []func(*Spec){
		func(s *Spec) { s.Criteria = nil },
		func(s *Spec) { s.Observation.CriteriaInputs = nil; s.Observation.Inputs = []int64{1} },
		func(s *Spec) { s.Observation.OracleActivity = "Expected" },
		func(s *Spec) { s.Observation.Inputs = []int64{1} },
		func(s *Spec) { s.Observation.ValueInputs = []string{"[1]"} },
		func(s *Spec) { s.Observation.CriteriaInputs = []string{"[0]", "[-0]"} },
		func(s *Spec) { s.Observation.CriteriaInputs[0] = s.Criteria.HoldoutInputs[0] },
		func(s *Spec) { s.Observation.CriteriaInputs[0] = "{}" },
	} {
		copy := s.Clone()
		mutate(copy)
		if copy.Validate() == nil {
			t.Fatal("invalid criteria question accepted", copy)
		}
	}
}
