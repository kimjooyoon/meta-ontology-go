package assemblyspec

import (
	"strings"
	"testing"
)

func TestCriteriaStrategyIsSourceOwnedAndSpecificToQuestions(t *testing.T) {
	s := criteriaSpec()
	s.Observation = &Observation{CriteriaInputs: []string{"[0]", "[-7]"}, MaxCandidates: 2, MaxRounds: 1}
	before, _ := s.Canonical()
	if strings.Contains(before, "criteria_strategy") {
		t.Fatal("legacy canonical form changed")
	}
	s.Observation.CriteriaStrategy = "distinguish_then_check"
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
	encoded, _ := s.Canonical()
	decoded, err := DecodeCanonical(encoded)
	if err != nil || encoded == before || decoded.Observation.CriteriaStrategy != s.Observation.CriteriaStrategy {
		t.Fatal(err, decoded)
	}
	for _, bad := range []string{"random", "distinguish"} {
		copy := s.Clone()
		copy.Observation.CriteriaStrategy = bad
		if copy.Validate() == nil {
			t.Fatal("unknown strategy accepted", bad)
		}
	}
	s.Observation.CriteriaInputs = nil
	s.Observation.Inputs = []int64{1}
	if s.Validate() == nil {
		t.Fatal("criteria strategy silently applied to scalar observation")
	}
}
