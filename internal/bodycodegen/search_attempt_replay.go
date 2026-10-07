package bodycodegen

import (
	"context"
	"fmt"
	"reflect"
)

// Replay the finite observations in their recorded order. This does not attest
// historical provider calls, timings or the origin of the recorded ordering.
func replayIRSearchAttempts(ctx context.Context, pkg, body string, prior Result, plan IRBodySearchPlan) error {
	s := prior.Report.BodySearch
	if len(s.Attempts) == 0 || len(s.Attempts) > min(plan.MaxAttempts, len(plan.Candidates)) {
		return fmt.Errorf("IR search attempt count exceeds its source contract")
	}
	seen := make(map[string]bool, len(s.Attempts))
	best, selected, evaluated := -1, "", 0
	for i, a := range s.Attempts {
		candidate, exists := candidateByID(plan.Candidates, a.CandidateID)
		if !exists || seen[a.CandidateID] || candidate.Expression != a.Expression {
			return fmt.Errorf("IR search attempt %d has an unknown, repeated or changed candidate", i)
		}
		seen[a.CandidateID] = true
		if err := replayIRSearchAttempt(ctx, pkg, body, prior.Report, plan, a); err != nil {
			return fmt.Errorf("IR search attempt %d: %w", i, err)
		}
		if a.ScoringCompleted {
			evaluated++
			if a.TestCasesPassed > best {
				best, selected = a.TestCasesPassed, a.CandidateID
			}
			if best == len(plan.TestCases) && i != len(s.Attempts)-1 {
				return fmt.Errorf("IR search continued after its first complete training result")
			}
		}
	}
	return replayIRSearchAttemptTotals(s, plan, best, selected, evaluated)
}

func replayIRSearchAttempt(ctx context.Context, pkg, body string, report Report,
	plan IRBodySearchPlan, a IRBodySearchAttempt) error {
	_, results, passed, typed, evaluationErr := evaluateSearchCandidate(ctx, pkg, report.Activity,
		report.ActivityID, body, bodyFillHoleToken(plan.HoleID), a.Expression, plan.TestCases)
	if err := ctx.Err(); err != nil {
		return err
	}
	failure := ""
	var accuracy *float64
	if evaluationErr != nil {
		failure = evaluationErr.Error()
		results, passed = nil, 0
	} else {
		value := float64(passed) * 100 / float64(len(plan.TestCases))
		accuracy = &value
	}
	if a.TypecheckPassed != typed || a.ScoringCompleted != (evaluationErr == nil) || a.Error != failure ||
		a.TestCasesPassed != passed || a.TestCasesTotal != len(plan.TestCases) ||
		!reflect.DeepEqual(a.AccuracyPercent, accuracy) || !equalIRBodyFillCaseResults(a.CaseResults, results) {
		return fmt.Errorf("finite observations differ from the source-declared cases")
	}
	return nil
}

func replayIRSearchAttemptTotals(s *IRBodySearchReceipt, plan IRBodySearchPlan, best int, selected string, evaluated int) error {
	stop := "MAX_ATTEMPTS"
	if len(s.Attempts) == len(plan.Candidates) {
		stop = "CANDIDATE_SET_EXHAUSTED"
	}
	if best == len(plan.TestCases) {
		stop = "TRAINING_SUITE_PASSED"
	} else if len(s.Attempts) < min(plan.MaxAttempts, len(plan.Candidates)) {
		return fmt.Errorf("IR search stopped before its declared budget without a complete training result")
	}
	accuracy := float64(best) * 100 / float64(len(plan.TestCases))
	if best < 0 || selected != s.SelectedCandidateID || best != s.TrainingPassed ||
		s.AttemptedCandidates != len(s.Attempts) || s.EvaluatedCandidates != evaluated ||
		s.UntestedCandidates != len(plan.Candidates)-len(s.Attempts) || s.StopReason != stop ||
		!reflect.DeepEqual(s.BestObservedAccuracyPercent, &accuracy) || s.GlobalBestAccuracyPercent != nil {
		return fmt.Errorf("IR search attempt totals or selected best candidate differ")
	}
	return nil
}
