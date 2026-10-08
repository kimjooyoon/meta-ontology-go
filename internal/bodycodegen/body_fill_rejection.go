package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// IRBodyFillCandidateRejection records a checked assignment without inventing a
// functional score. Only candidates in CandidateScores completed local scoring.
type IRBodyFillCandidateRejection struct {
	CandidateID string               `json:"candidate_id"`
	Stage       string               `json:"stage"`
	Reason      string               `json:"reason"`
	HoleFills   []IRBodyFillHoleFill `json:"hole_fills"`
}

type IRBodyFillFailure struct {
	Activity   string                         `json:"activity"`
	PlanSHA256 string                         `json:"plan_sha256"`
	Rejected   []IRBodyFillCandidateRejection `json:"rejected_candidates"`
}

// NoValidBodyFillCandidates is terminal and preserves every rejected assignment.
type NoValidBodyFillCandidates struct{ Observation IRBodyFillFailure }

func (e *NoValidBodyFillCandidates) Error() string {
	var reasons []string
	for _, r := range e.Observation.Rejected {
		reasons = append(reasons, r.CandidateID+" ("+r.Stage+"): "+r.Reason)
	}
	return fmt.Sprintf("activity %s has no valid body-fill assignment: %s", e.Observation.Activity, strings.Join(reasons, "; "))
}

func rejectedBodyFillCandidate(plan IRBodyFillPlan, candidate IRBodyFillCandidate, stage string, err error) IRBodyFillCandidateRejection {
	return IRBodyFillCandidateRejection{CandidateID: candidate.ID, Stage: stage, Reason: err.Error(),
		HoleFills: bodyFillHoleResults(bodyFillPlanHoles(plan), bodyFillCandidateFills(plan, candidate))}
}

func noValidBodyFill(ctx context.Context, activity string, plan IRBodyFillPlan, rejected []IRBodyFillCandidateRejection) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	raw, err := json.Marshal(plan)
	if err != nil {
		return err
	}
	return &NoValidBodyFillCandidates{Observation: IRBodyFillFailure{Activity: activity, PlanSHA256: digest(raw), Rejected: rejected}}
}

// SourceFillCandidateRejection is limited to deterministic type/training errors.
// Cancellation and separate holdout evaluation failures remain ordinary errors.
type SourceFillCandidateRejection struct{ Observation IRBodyFillCandidateRejection }

func (e *SourceFillCandidateRejection) Error() string { return e.Observation.Reason }
