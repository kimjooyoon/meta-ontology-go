package bodycodegen

import (
	"context"
	"fmt"
)

// RecordCandidate is a checked explicit combination, rather than the winner of
// the activity's local search. A caller-level search owns the choice of mask.
// The observation keeps the original local obligations and their actual values.
type RecordCandidate struct {
	Schema               string                `json:"schema"`
	Activity             string                `json:"activity"`
	ActivityID           string                `json:"activity_id"`
	InputSourceSHA256    string                `json:"input_source_sha256"`
	SelectedSourceSHA256 string                `json:"selected_source_sha256"`
	ContractSHA256       string                `json:"contract_sha256"`
	AttemptBudget        int                   `json:"attempt_budget"`
	CandidateCount       int                   `json:"candidate_count"`
	Attempt              RecordAssemblyAttempt `json:"attempt"`
	Cases                []RecordAssemblyCase  `json:"cases"`
}

// RealizeRecordCandidate checks one source-declared combination and makes its
// body callable. It does not change local expectations, claim a local optimum,
// load a model or write a legacy assembly receipt with different semantics.
func RealizeRecordCandidate(ctx context.Context, filename string, source []byte,
	activity string, mask uint16) (RecordCandidate, []byte, error) {
	p, err := prepareRecordAssembly(ctx, filename, source, activity)
	if err != nil {
		return RecordCandidate{}, nil, err
	}
	if int(mask) >= 1<<len(p.choices) {
		return RecordCandidate{}, nil, fmt.Errorf("candidate mask exceeds source choices")
	}
	attempt, cases, err := evaluateRecordMask(ctx, p, mask)
	if err != nil {
		return RecordCandidate{}, nil, err
	}
	if attempt.Total == 0 {
		return RecordCandidate{}, nil, fmt.Errorf("record candidate has no evaluable local obligations: %s", attempt.Reason)
	}
	r := newRecordAssemblyReceipt(source, p)
	r.SelectedMask, r.Attempts, r.Cases = mask, []RecordAssemblyAttempt{attempt}, cases
	r.Passed, r.Total = attempt.Passed, attempt.Total
	r.FieldsPassed, r.FieldsTotal = attempt.FieldsPassed, attempt.FieldsTotal
	finishRecordAssemblySearch(r)
	generation, err := emitRecordAssembly(ctx, filename, source, p, r)
	if err != nil {
		return RecordCandidate{}, nil, err
	}
	selected, err := fixedCalledSource(filename, generation.GoooSource, activity)
	if err != nil {
		return RecordCandidate{}, nil, err
	}
	return RecordCandidate{Schema: "gooo/record-candidate/v1", Activity: activity, ActivityID: p.body.activityID,
		InputSourceSHA256: digest(source), SelectedSourceSHA256: digest([]byte(selected)),
		ContractSHA256: r.ContractSHA256, AttemptBudget: p.spec.MaxAttempts,
		CandidateCount: len(r.Ranking), Attempt: attempt, Cases: cases}, []byte(selected), nil
}
