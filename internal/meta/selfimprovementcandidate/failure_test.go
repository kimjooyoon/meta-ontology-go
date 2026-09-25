package selfimprovementcandidate

import "testing"

func TestEvaluateClosesSemanticCounterexamples(t *testing.T) {
	head, runID := fixtureSHA("c"), int64(44)
	tests := []struct {
		name, reason, resolution string
		mutate                   func(*sourceObservation)
		reseal                   bool
	}{
		{"unknown decision", ReasonSourceUnknown, ResolutionLower, func(s *sourceObservation) { s.Decision = "MAYBE" }, true},
		{"lowered source", ReasonSourceLowered, ResolutionLower, func(s *sourceObservation) { s.Resolution = ResolutionLower }, true},
		{"explicit failure", ReasonSourceRejected, ResolutionExact, func(s *sourceObservation) { s.Decision = "FAIL_CLOSED" }, true},
		{"head drift", ReasonSourceIdentity, ResolutionExact, func(s *sourceObservation) { s.SubjectSHA = fixtureSHA("d") }, true},
		{"digest mismatch", ReasonSourceIntegrity, ResolutionExact, func(s *sourceObservation) { s.Digest = zeroDigest }, false},
		{"candidate leak", ReasonSourceCandidate, ResolutionExact, func(s *sourceObservation) { s.Summary.CandidateCount = 1 }, true},
		{"authority leak", ReasonSourceAuthority, ResolutionExact, func(s *sourceObservation) { s.Authority.ExecutionAuthorized = true }, true},
		{"gap absent", ReasonGapAbsent, ResolutionExact, func(s *sourceObservation) { s.NotClaimed = s.NotClaimed[2:] }, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source := validSource(head, runID)
			test.mutate(&source)
			if test.reseal {
				reseal(&source)
			}
			report := Evaluate(validRepository(), candidateContractPath, head, runID, sourceBytes(source))
			if report.Decision != DecisionFailClosed || report.Reason != test.reason ||
				report.Resolution != test.resolution || len(report.Candidates) != 0 {
				t.Fatalf("counterexample escaped: %+v", report)
			}
			if err := Validate(report, head, runID); err != nil {
				t.Fatal(err)
			}
		})
	}
}
