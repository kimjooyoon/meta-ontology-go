package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
)

// RetainedSearchCandidateSet stores the bounded expressions without duplicating
// the potentially large hole-context observations already held by the receipt.
type RetainedSearchCandidateSet struct {
	PlanSHA256       string                `json:"plan_sha256"`
	AttemptBudget    int                   `json:"attempt_budget"`
	Candidates       []IRBodyFillCandidate `json:"candidates"`
	GenerationSHA256 string                `json:"generation_sha256"`
}

func retainedSearchCandidateSet(plan IRBodySearchPlan, planSHA string,
	generation *IRBodySearchCandidateGenerationReceipt) *RetainedSearchCandidateSet {
	raw, _ := json.Marshal(generation)
	return &RetainedSearchCandidateSet{PlanSHA256: planSHA, AttemptBudget: plan.MaxAttempts,
		Candidates: slices.Clone(plan.Candidates), GenerationSHA256: digest(raw)}
}

// RetainedSourceSearchCandidates reads the bounded set of an initial generation
// already constructed or source-replayed by the caller. The current activity
// must keep that exact source contract; called helper bodies may have changed.
// VerifyIRBodySearchProjection separately verifies the set's derivation against
// its original source checkpoint. Reading this set runs no helper or model.
func RetainedSourceSearchCandidates(ctx context.Context, filename string, source []byte, activity string,
	prior *IRBodySearchReceipt) (SourceSearchCandidateSet, error) {
	_, set, err := retainedSourceSearchPlan(ctx, filename, source, activity, prior)
	return set, err
}

func retainedSourceSearchPlan(ctx context.Context, filename string, source []byte, activity string,
	prior *IRBodySearchReceipt) (IRBodySearchPlan, SourceSearchCandidateSet, error) {
	var set SourceSearchCandidateSet
	if ctx == nil || prior == nil || prior.CandidateSet == nil || prior.AttemptBudget == nil || prior.OriginalSourceDigest == "" {
		return IRBodySearchPlan{}, set, fmt.Errorf("retained search candidates require a context and initial source receipt")
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
	if err := validateIRBodySearchPlan(plan); err != nil {
		return plan, set, err
	}
	snapshot := prior.CandidateSet
	set = SourceSearchCandidateSet{PlanSHA256: snapshot.PlanSHA256, AttemptBudget: snapshot.AttemptBudget,
		Candidates: slices.Clone(snapshot.Candidates), Generation: prior.CandidateGeneration}
	generation, _ := json.Marshal(prior.CandidateGeneration)
	if len(set.Candidates) < 2 || len(set.Candidates) > plan.CandidateGeneration.MaxCandidates ||
		set.AttemptBudget != plan.MaxAttempts || *prior.AttemptBudget != plan.MaxAttempts ||
		prior.CandidateCount != len(set.Candidates) || prior.IRPlanSHA256 != set.PlanSHA256 ||
		snapshot.GenerationSHA256 != digest(generation) {
		return plan, set, fmt.Errorf("retained search candidate bounds differ from the initial contract")
	}
	plan.Candidates = set.Candidates
	raw, err := json.Marshal(plan)
	if err != nil || digest(raw) != set.PlanSHA256 {
		return plan, set, fmt.Errorf("retained search candidate plan differs from the source contract")
	}
	return plan, set, nil
}

// RealizeRetainedSourceSearchCandidate applies one expression from an initial
// source-verified candidate set to the current helper checkpoint. The activity's
// cases and budget are unchanged. Its current executions, not the old scores,
// decide local admissibility. No context-dependent candidate pool is regenerated.
func RealizeRetainedSourceSearchCandidate(ctx context.Context, filename string, source []byte, activity string,
	prior *IRBodySearchReceipt, candidateID string) (SearchCandidate, []byte, error) {
	plan, set, err := retainedSourceSearchPlan(ctx, filename, source, activity, prior)
	if err != nil {
		return SearchCandidate{}, nil, err
	}
	ctx, calls := captureSearchCalls(ctx, filename, source, activity)
	return realizeSourceSearchCandidatePlan(ctx, filename, source, activity, candidateID, plan, set, calls, prior.OriginalSourceDigest)
}
