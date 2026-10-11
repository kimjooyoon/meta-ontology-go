package assemblyspec

import (
	"reflect"
	"testing"
)

func TestConditionCaseContractCarrierAndLimits(t *testing.T) {
	s := validSpec()
	s.ConditionCases = []ConditionCase{{ChoiceID: "offset", Input: 9007199254740993, Expected: false}}
	raw, err := s.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeCanonical(raw)
	if err != nil || !reflect.DeepEqual(&s, decoded) {
		t.Fatal("condition carrier changed", err)
	}
	clone := s.Clone()
	clone.ConditionCases[0].Input++
	if s.ConditionCases[0].Input != 9007199254740993 {
		t.Fatal("clone shares condition cases")
	}
	for _, change := range []func(*Spec){
		func(s *Spec) { s.ConditionCases = append(s.ConditionCases, s.ConditionCases[0]) },
		func(s *Spec) { s.ConditionCases[0].ChoiceID = "missing" },
		func(s *Spec) { s.ConditionCases = make([]ConditionCase, 129) },
		func(s *Spec) { s.Search = &Search{} },
		func(s *Spec) { s.FillPlan = &FillPlan{} },
		func(s *Spec) { s.ValueCases = []ValueCase{{}} },
		func(s *Spec) { s.Choices[0].Kind = "root_order" },
	} {
		changed := s.Clone()
		change(changed)
		if err := changed.Validate(); err == nil {
			t.Fatal("invalid condition contract accepted", changed)
		}
	}
}
