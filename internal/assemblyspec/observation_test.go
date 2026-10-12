package assemblyspec

import "testing"

func TestSourceObserveSpecKeepsCanonicalOwnedMeaning(t *testing.T) {
	s := validSpec()
	s.Observation = &Observation{Inputs: []int64{-9007199254740995, 7}, OracleActivity: "기준",
		MaxCandidates: 2, MaxRounds: 1, ReuseProbeOutputs: true, ResolveUniqueCandidate: true}
	raw, err := s.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeCanonical(raw)
	if err != nil {
		t.Fatal(err)
	}
	clone := decoded.Clone()
	clone.Observation.Inputs[0], clone.Observation.MaxCandidates = 0, 1
	if decoded.Observation.Inputs[0] != -9007199254740995 || decoded.Observation.MaxCandidates != 2 {
		t.Fatal("clone changed exact source questions or budget")
	}
	for _, mutate := range []func(*Spec){
		func(s *Spec) { s.Observation.Inputs = nil },
		func(s *Spec) { s.Observation.Inputs = make([]int64, 33) },
		func(s *Spec) { s.Observation.Inputs = []int64{7, 7} },
		func(s *Spec) { s.Observation.MaxCandidates = 65 },
		func(s *Spec) { s.Observation.MaxRounds = 9 },
		func(s *Spec) { s.Observation.OracleActivity = "a.b" },
		func(s *Spec) { s.Search = &Search{} },
		func(s *Spec) { s.FillPlan = &FillPlan{} },
		func(s *Spec) { s.Cases = nil },
		func(s *Spec) { s.HoldoutCases = []Case{{Input: 7}} },
		func(s *Spec) { s.ValueCases = []ValueCase{{Inputs: "[]", Expected: "{}"}} },
	} {
		changed := decoded.Clone()
		mutate(changed)
		if err := changed.Validate(); err == nil {
			t.Fatal("invalid source observation accepted", changed)
		}
	}
}
