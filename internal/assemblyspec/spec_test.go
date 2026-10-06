package assemblyspec

import (
	"strings"
	"testing"
)

func validSpec() Spec {
	return Spec{Choices: []Choice{{ID: "offset", Kind: "operand_order", Intent: "2 minus input"}},
		Cases: []Case{{Input: 4, Expected: -2}}, MaxAttempts: 2}
}

func TestCanonicalAssemblyCarrierAndOwnedStorage(t *testing.T) {
	s := validSpec()
	raw, err := s.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	restored, err := DecodeCanonical(raw)
	if err != nil {
		t.Fatal(err)
	}
	restored.Choices[0].Intent = "changed"
	restored.Cases[0].Expected = 0
	if s.Choices[0].Intent == "changed" || s.Cases[0].Expected == 0 {
		t.Fatal("decoded assembly shares caller storage")
	}
	for _, invalid := range []string{"", " " + raw, raw + " ", strings.Replace(raw, `"choices":`, `"Choices":`, 1),
		strings.TrimSuffix(raw, "}") + `,"unknown":1}`, strings.TrimSuffix(raw, "}") + `,"max_attempts":2}`} {
		if _, err := DecodeCanonical(invalid); err == nil {
			t.Fatal("noncanonical assembly carrier accepted", invalid)
		}
	}
}

func TestAssemblyFiniteAndTextBudgets(t *testing.T) {
	for _, mutate := range []func(*Spec){
		func(s *Spec) { s.Choices = nil }, func(s *Spec) { s.Cases = nil },
		func(s *Spec) { s.MaxAttempts = 0 }, func(s *Spec) { s.MaxAttempts = 65 },
		func(s *Spec) { s.Seed = strings.Repeat("x", 513) },
		func(s *Spec) { s.Choices[0].ID = "" },
		func(s *Spec) { s.Choices[0].Intent = strings.Repeat("x", 513) },
		func(s *Spec) { s.Choices[0].Kind = "call" },
		func(s *Spec) { s.Choices[0].Occurrence = 128 },
		func(s *Spec) { s.Choices[0].Alternative = "x" },
		func(s *Spec) { s.Choices[0].Kind = "assignment_target" },
		func(s *Spec) { s.Choices = append(s.Choices, s.Choices[0]) },
	} {
		s := validSpec()
		mutate(&s)
		if err := s.Validate(); err == nil {
			t.Fatal("invalid assembly budget accepted", s)
		}
	}
	s := validSpec()
	s.Choices[0].Kind, s.Choices[0].Alternative = "assignment_target", "value"
	if err := s.Validate(); err != nil {
		t.Fatal("supported name path rejected", err)
	}
}

func TestRecordSourceFillAllowsOnlyTypedRecordGrammars(t *testing.T) {
	spec := Spec{
		ValueCases: []ValueCase{
			{Inputs: `[{"state":"ready"}]`, Expected: `{"decision":"accepted"}`},
			{Inputs: `[{"state":"queued"}]`, Expected: `{"decision":"rejected"}`},
		},
		FillPlan: &FillPlan{
			Intent: "route record states",
			Holes:  []FillHole{{ID: "condition"}, {ID: "accepted"}, {ID: "rejected"}},
			Generation: &FillGeneration{MaxCandidates: 16, HoleGrammars: []FillHoleGrammar{
				{HoleID: "condition", Grammar: "record-field-predicate/v1", MaxExpressions: 4},
				{HoleID: "accepted", Grammar: "record-string-literal/v1", MaxExpressions: 2},
				{HoleID: "rejected", Grammar: "record-string-literal/v1", MaxExpressions: 2},
			}},
		},
	}
	if err := spec.Validate(); err != nil {
		t.Fatalf("typed record candidate derivation rejected: %v", err)
	}
	composed := spec.Clone()
	composed.FillPlan.Generation.HoleGrammars[0].Grammar = "record-field-predicate-composition/v1"
	if err := composed.Validate(); err != nil {
		t.Fatalf("composed record predicate derivation rejected: %v", err)
	}
	ordered := spec.Clone()
	ordered.FillPlan.Generation.HoleGrammars[0].Grammar = "record-field-predicate/v2"
	if err := ordered.Validate(); err != nil {
		t.Fatalf("ordered record predicate derivation rejected: %v", err)
	}
	relation := spec.Clone()
	relation.FillPlan.Generation.HoleGrammars[0].Grammar = "record-field-relation/v1"
	if err := relation.Validate(); err != nil {
		t.Fatalf("record field relation derivation rejected: %v", err)
	}
	orderedComposed := spec.Clone()
	orderedComposed.FillPlan.Generation.HoleGrammars[0].Grammar = "record-field-predicate-composition/v2"
	if err := orderedComposed.Validate(); err != nil {
		t.Fatalf("ordered composed record predicate derivation rejected: %v", err)
	}

	invalid := spec.Clone()
	invalid.FillPlan.Generation.HoleGrammars[0].Grammar = "integer-predicate/v1"
	if err := invalid.Validate(); err == nil {
		t.Fatal("integer grammar mixed into record value cases")
	}
}

func TestRecordSourceFillHoldoutsAreCanonicalDisjointAndCloned(t *testing.T) {
	spec := Spec{
		ValueCases:        []ValueCase{{Inputs: `[{"state":"ready"}]`, Expected: `{"decision":"yes"}`}},
		ValueHoldoutCases: []ValueCase{{Inputs: `[{"state":"queued"}]`, Expected: `{"decision":"no"}`}},
		FillPlan: &FillPlan{Intent: "route by state", Holes: []FillHole{{ID: "condition"}, {ID: "yes"}},
			Candidates: []FillCandidate{{ID: "a", Fills: []HoleFilling{{HoleID: "condition", Expression: `input.state == "ready"`}, {HoleID: "yes", Expression: `"yes"`}}},
				{ID: "b", Fills: []HoleFilling{{HoleID: "condition", Expression: `input.state != "ready"`}, {HoleID: "yes", Expression: `"no"`}}}}},
	}
	if err := spec.Validate(); err != nil {
		t.Fatalf("disjoint record holdout rejected: %v", err)
	}
	clone := spec.Clone()
	clone.ValueHoldoutCases[0].Expected = `{"decision":"changed"}`
	if spec.ValueHoldoutCases[0].Expected != `{"decision":"no"}` {
		t.Fatal("record holdout clone shares caller storage")
	}
	for _, mutate := range []func(*Spec){
		func(s *Spec) { s.ValueHoldoutCases[0].Inputs = s.ValueCases[0].Inputs },
		func(s *Spec) { s.ValueHoldoutCases[0].Expected = `{"decision": "no"}` },
		func(s *Spec) { s.FillPlan = nil },
	} {
		invalid := spec.Clone()
		mutate(invalid)
		if err := invalid.Validate(); err == nil {
			t.Fatal("invalid record holdout accepted", invalid)
		}
	}
}

func TestAssemblyCheckpointCanonicalSelectionAndClone(t *testing.T) {
	s := validSpec()
	s.Baseline = "return input - 2"
	s.Picked = []Pick{{ID: "offset", Label: "layout_reverse"}}
	raw, err := s.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeCanonical(raw)
	if err != nil || decoded.Baseline != s.Baseline {
		t.Fatal(err)
	}
	clone := s.Clone()
	clone.Picked[0].Label = "layout_forward"
	if s.Picked[0].Label != "layout_reverse" {
		t.Fatal("checkpoint clone shares picked storage")
	}
	for _, mutate := range []func(*Spec){
		func(s *Spec) { s.Baseline = "" }, func(s *Spec) { s.Picked = nil },
		func(s *Spec) { s.Picked[0].ID = "other" }, func(s *Spec) { s.Picked[0].Label = "reference_second" },
		func(s *Spec) { s.Picked = append(s.Picked, s.Picked[0]) },
		func(s *Spec) { s.Baseline = strings.Repeat("x", 1+(128<<10)) },
	} {
		changed := s.Clone()
		mutate(changed)
		if changed.Validate() == nil {
			t.Fatal("invalid checkpoint accepted", changed)
		}
	}
}
