package bodycodegen

import (
	"context"
	"testing"
)

func TestSourceIRSearchReplayRecomputesEachAttempt(t *testing.T) {
	changes := map[string]func(*IRBodySearchReceipt){
		"score":                     func(s *IRBodySearchReceipt) { s.Attempts[0].TestCasesPassed++ },
		"unscored":                  func(s *IRBodySearchReceipt) { s.Attempts[0].ScoringCompleted = false },
		"type":                      func(s *IRBodySearchReceipt) { s.Attempts[0].TypecheckPassed = false },
		"expression":                func(s *IRBodySearchReceipt) { s.Attempts[0].Expression = "42" },
		"case":                      func(s *IRBodySearchReceipt) { s.Attempts[0].CaseResults[0].Actual++ },
		"count":                     func(s *IRBodySearchReceipt) { s.AttemptedCandidates++ },
		"evaluated":                 func(s *IRBodySearchReceipt) { s.EvaluatedCandidates++ },
		"remaining":                 func(s *IRBodySearchReceipt) { s.UntestedCandidates++ },
		"stop":                      func(s *IRBodySearchReceipt) { s.StopReason = "MAX_ATTEMPTS" },
		"best":                      func(s *IRBodySearchReceipt) { v := 12.0; s.BestObservedAccuracyPercent = &v },
		"global":                    func(s *IRBodySearchReceipt) { v := 100.0; s.GlobalBestAccuracyPercent = &v },
		"duplicate":                 func(s *IRBodySearchReceipt) { s.Attempts = append([]IRBodySearchAttempt{s.Attempts[0]}, s.Attempts...) },
		"continue-after-completion": func(s *IRBodySearchReceipt) { s.Attempts = append(s.Attempts, s.Attempts[0]) },
		"missing":                   func(s *IRBodySearchReceipt) { s.Attempts = nil },
	}
	for name, change := range changes {
		t.Run(name, func(t *testing.T) {
			source, prior := sourceSearchReplayFixture(t)
			change(prior.Report.BodySearch)
			if err := VerifyIRBodySearchProjection(context.Background(), "search.gooo", source, prior); err == nil {
				t.Fatal("changed attempt history was accepted", name)
			}
		})
	}
}

func TestSourceIRSearchReplayCancellation(t *testing.T) {
	source, prior := sourceSearchReplayFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := VerifyIRBodySearchProjection(ctx, "search.gooo", source, prior); err == nil {
		t.Fatal("canceled replay continued")
	}
}
