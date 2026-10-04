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
