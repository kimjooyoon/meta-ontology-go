package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
)

// SourceSearchCandidateSet binds the grammar and its retained expressions to the
// source. AttemptBudget further bounds the prefix eligible for caller search.
type SourceSearchCandidateSet struct {
	PlanSHA256    string                                  `json:"plan_sha256"`
	AttemptBudget int                                     `json:"attempt_budget"`
	Candidates    []IRBodyFillCandidate                   `json:"candidates"`
	Generation    *IRBodySearchCandidateGenerationReceipt `json:"generation"`
}

type SearchCandidate struct {
	Schema               string                                  `json:"schema"`
	Activity             string                                  `json:"activity"`
	ActivityID           string                                  `json:"activity_id"`
	InputSourceSHA256    string                                  `json:"input_source_sha256"`
	SelectedSourceSHA256 string                                  `json:"selected_source_sha256"`
	PlanSHA256           string                                  `json:"plan_sha256"`
	AttemptBudget        int                                     `json:"attempt_budget"`
	CandidateCount       int                                     `json:"candidate_count"`
	Generation           *IRBodySearchCandidateGenerationReceipt `json:"candidate_generation"`
	Attempt              IRBodySearchAttempt                     `json:"attempt"`
}

func sourceSearchCandidatePlan(ctx context.Context, filename string, source []byte, activity string) (IRBodySearchPlan, SourceSearchCandidateSet, error) {
	var set SourceSearchCandidateSet
	if ctx == nil {
		return IRBodySearchPlan{}, set, fmt.Errorf("source search candidate requires a context")
	}
	if err := ctx.Err(); err != nil {
		return IRBodySearchPlan{}, set, err
	}
	spec, err := SourceAssembly(ctx, filename, source, activity)
	if err != nil {
		return IRBodySearchPlan{}, set, err
	}
	plan, err := sourceIRSearchPlan(spec)
	if err != nil {
		return plan, set, err
	}
	if err = validateIRBodySearchPlan(plan); err != nil {
		return plan, set, err
	}
	generation, err := generateIRBodySearchCandidatesForSource(ctx, filename, source, activity, &plan)
	if err != nil {
		return plan, set, err
	}
	raw, err := json.Marshal(plan)
	set = SourceSearchCandidateSet{PlanSHA256: digest(raw), AttemptBudget: plan.MaxAttempts, Candidates: plan.Candidates, Generation: generation}
	return plan, set, err
}

func PlanSourceSearchCandidates(ctx context.Context, filename string, source []byte, activity string) (SourceSearchCandidateSet, error) {
	_, set, err := sourceSearchCandidatePlan(ctx, filename, source, activity)
	return set, err
}

// RealizeSourceSearchCandidate checks a caller-chosen expression without
// weakening local cases or replacing a search winner's historical receipt.
// Holdout cases do not influence its admissibility or local selection score.
func RealizeSourceSearchCandidate(ctx context.Context, filename string, source []byte, activity, candidateID string) (SearchCandidate, []byte, error) {
	plan, set, err := sourceSearchCandidatePlan(ctx, filename, source, activity)
	if err != nil {
		return SearchCandidate{}, nil, err
	}
	candidate, found := candidateByID(plan.Candidates[:min(plan.MaxAttempts, len(plan.Candidates))], candidateID)
	if !found {
		return SearchCandidate{}, nil, fmt.Errorf("search candidate exceeds the source grammar or attempt budget")
	}
	file, declaration, id, body, err := prepareBodySearch(filename, source, activity, plan.HoleID)
	if err != nil {
		return SearchCandidate{}, nil, err
	}
	filled, cases, passed, typed, err := evaluateSearchCandidate(ctx, file.Package.Name, activity, id, body, bodyFillHoleToken(plan.HoleID), candidate.Expression, plan.TestCases)
	if err != nil {
		return SearchCandidate{}, nil, err
	}
	selected, err := replaceActivityProgram(source, declaration.ValueProgramSpan, filled)
	if err != nil {
		return SearchCandidate{}, nil, err
	}
	fixed, err := fixedCalledSource(filename, string(selected), activity)
	if err != nil {
		return SearchCandidate{}, nil, err
	}
	if _, err = GenerateWithPlanner(ctx, filename, []byte(fixed), activity, "", ""); err != nil {
		return SearchCandidate{}, nil, err
	}
	accuracy := float64(passed) * 100 / float64(len(plan.TestCases))
	observation := SearchCandidate{Schema: "gooo/search-candidate/v1", Activity: activity, ActivityID: id,
		InputSourceSHA256: digest(source), SelectedSourceSHA256: digest([]byte(fixed)), PlanSHA256: set.PlanSHA256,
		AttemptBudget: set.AttemptBudget, CandidateCount: len(set.Candidates), Generation: set.Generation,
		Attempt: IRBodySearchAttempt{CandidateID: candidate.ID, Expression: candidate.Expression, SelectionMethod: "caller_selected_candidate", TypecheckPassed: typed,
			ScoringCompleted: true, TestCasesPassed: passed, TestCasesTotal: len(plan.TestCases), AccuracyPercent: &accuracy, CaseResults: cases}}
	return observation, []byte(fixed), nil
}
