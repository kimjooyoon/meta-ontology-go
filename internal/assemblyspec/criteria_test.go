package assemblyspec

import (
	"reflect"
	"testing"
)

func criteriaSpec() Spec {
	return Spec{Choices: []Choice{{ID: "amount", Kind: "field_update", Alternative: "0 - input.amount", Intent: "Negate."}},
		MaxAttempts: 2, Criteria: &Criteria{Predicates: []Criterion{{ID: "amount", Activity: "AmountSatisfied"}},
			Inputs: []string{"[9007199254740993]"}, HoldoutInputs: []string{"[9007199254740995]"}}}
}

func TestCriteriaCanonicalCloneAndIndependentCounts(t *testing.T) {
	s := criteriaSpec()
	canonical, err := s.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeCanonical(canonical)
	if err != nil || !reflect.DeepEqual(decoded, &s) {
		t.Fatal("criteria did not survive canonical carrier", err)
	}
	clone := s.Clone()
	clone.Criteria.Predicates[0].Activity = "Changed"
	clone.Criteria.Inputs[0], clone.Criteria.HoldoutInputs[0] = "[1]", "[2]"
	if s.Criteria.Predicates[0].Activity != "AmountSatisfied" || s.Criteria.Inputs[0] != "[9007199254740993]" ||
		s.Criteria.HoldoutInputs[0] != "[9007199254740995]" || len(s.ValueCases) != 0 {
		t.Fatal("clone aliases criteria or manufactures exact-result expectations")
	}
}

func TestCriteriaRejectDuplicateNamesAndLeakingInputs(t *testing.T) {
	changes := []func(*Spec){
		func(s *Spec) { s.Criteria.Predicates = nil },
		func(s *Spec) { s.Criteria.Predicates = append(s.Criteria.Predicates, s.Criteria.Predicates[0]) },
		func(s *Spec) { s.Criteria.Predicates[0].Activity = "" },
		func(s *Spec) { s.Criteria.Inputs = nil },
		func(s *Spec) { s.Criteria.Inputs = []string{"[0]", "[-0]"} },
		func(s *Spec) { s.Criteria.Inputs[0] = "[1.5]" },
		func(s *Spec) { s.Criteria.Inputs[0] = "[]" },
		func(s *Spec) { s.Criteria.Inputs[0] = "{}" },
		func(s *Spec) { s.Criteria.HoldoutInputs = append(s.Criteria.HoldoutInputs, s.Criteria.Inputs[0]) },
		func(s *Spec) { s.ValueCases = []ValueCase{{Inputs: s.Criteria.HoldoutInputs[0], Expected: "{}"}} },
		func(s *Spec) { s.ValueHoldoutCases = []ValueCase{{Inputs: s.Criteria.Inputs[0], Expected: "{}"}} },
		func(s *Spec) { s.Choices[0].Kind = "local_reference" },
	}
	for i, change := range changes {
		s := criteriaSpec()
		change(&s)
		if err := s.Validate(); err == nil {
			t.Fatalf("invalid criteria %d accepted", i)
		}
	}
}
