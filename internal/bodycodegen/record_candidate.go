package bodycodegen

import (
	"context"
	"fmt"
)

// RecordCandidate is a checked explicit combination, rather than the winner of
// the activity's local search. A caller-level search owns the choice of mask.
// The observation keeps the original local obligations and their actual values.
type RecordCandidate struct {
	CriteriaQuestions    *CriteriaQuestionRetention  `json:"criteria_questions,omitempty"`
	Observation          *RecordCandidateObservation `json:"observation,omitempty"`
	Schema               string                      `json:"schema"`
	Activity             string                      `json:"activity"`
	ActivityID           string                      `json:"activity_id"`
	InputSourceSHA256    string                      `json:"input_source_sha256"`
	SelectedSourceSHA256 string                      `json:"selected_source_sha256"`
	ContractSHA256       string                      `json:"contract_sha256"`
	AttemptBudget        int                         `json:"attempt_budget"`
	CandidateCount       int                         `json:"candidate_count"`
	Attempt              RecordAssemblyAttempt       `json:"attempt"`
	Cases                []RecordAssemblyCase        `json:"cases"`
}

// SourceRecordCandidateRejection is a typed or evaluated source candidate that
// cannot satisfy an evaluable local contract. Cancellation is not a rejection.
type SourceRecordCandidateRejection struct{ cause error }

func (e *SourceRecordCandidateRejection) Error() string { return e.cause.Error() }
func (e *SourceRecordCandidateRejection) Unwrap() error { return e.cause }

// RealizeRecordCandidate checks one source-declared combination and makes its
// body callable. It does not change local expectations, claim a local optimum,
// load a model or write a legacy assembly receipt with different semantics.
func RealizeRecordCandidate(ctx context.Context, filename string, source []byte,
	activity string, mask uint16) (RecordCandidate, []byte, error) {
	return RealizeObservedRecordCandidate(ctx, filename, source, activity, nil, mask)
}

// RealizeObservedRecordCandidate carries a verified construction receipt into
// an explicit caller choice. It preserves acquired local answers across slots.
func RealizeObservedRecordCandidate(ctx context.Context, filename string, source []byte,
	activity string, prior *RecordAssemblyReceipt, mask uint16) (RecordCandidate, []byte, error) {
	p, err := prepareRecordAssembly(ctx, filename, source, activity)
	if err != nil {
		return RecordCandidate{}, nil, err
	}
	if int(mask) >= 1<<len(p.choices) {
		return RecordCandidate{}, nil, fmt.Errorf("candidate mask exceeds source choices")
	}
	questions, err := retainCriteriaQuestions(&p, prior)
	if err != nil {
		return RecordCandidate{}, nil, err
	}
	retained, err := retainRecordCandidateObservation(ctx, filename, source, &p, prior)
	if err != nil {
		return RecordCandidate{}, nil, err
	}
	attempt, cases, err := evaluateRecordMask(ctx, p, mask)
	if err != nil {
		return RecordCandidate{}, nil, err
	}
	r := newRecordAssemblyReceipt(source, p)
	observation := RecordCandidate{Schema: "gooo/record-candidate/v1", Activity: activity, ActivityID: p.body.activityID,
		InputSourceSHA256: digest(source), ContractSHA256: r.ContractSHA256, AttemptBudget: p.spec.MaxAttempts,
		CandidateCount: len(r.Ranking), Attempt: attempt, Cases: cases}
	if retained != nil {
		observation.Schema, observation.Observation = "gooo/record-candidate/v2", retained
	}
	if questions != nil {
		observation.Schema = "gooo/record-candidate/v3"
		observation.CriteriaQuestions = questions
	}
	if p.scalar != nil {
		observation.Schema = "gooo/scalar-criteria-candidate/v1"
	}
	if !usableRecordAttempt(attempt) {
		return observation, nil, &SourceRecordCandidateRejection{
			cause: fmt.Errorf("record candidate has no evaluable local obligations: %s", attempt.Reason)}
	}
	r.Attempts = []RecordAssemblyAttempt{attempt}
	selectRecordAttempt(r, attempt, cases)
	finishRecordAssemblySearch(r)
	generation, err := emitRecordAssembly(ctx, filename, source, p, r)
	if err != nil {
		return RecordCandidate{}, nil, err
	}
	selected, err := fixedCalledSource(filename, generation.GoooSource, activity)
	if err != nil {
		return RecordCandidate{}, nil, err
	}
	observation.SelectedSourceSHA256 = digest([]byte(selected))
	return observation, []byte(selected), nil
}
