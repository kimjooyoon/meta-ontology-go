package valueexecution

import (
	"encoding/json"
	"fmt"
)

const AcceptedRevisionNextRunStageSchema = "gooo/value-execution-accepted-revision-next-run-stage/v1"

const (
	AcceptedRevisionNextRunStageDecision = "STAGED"
	AcceptedRevisionNextRunStageReason   = "SOURCE_REVISION_STAGED_FOR_NEXT_RUN"
	AcceptedRevisionNextRunStageNextOp   = "EXECUTE_STAGED_ACCEPTED_REVISION"
)

// AcceptedRevisionNextRunStage is the caller-owned boundary between an
// accepted comparison and a later run. It records identity without granting
// permission to modify the source repository.
type AcceptedRevisionNextRunStage struct {
	Schema                string      `json:"schema"`
	Decision              string      `json:"decision"`
	State                 ReplayState `json:"state"`
	Outcome               string      `json:"outcome"`
	Reason                string      `json:"reason"`
	NextOperation         string      `json:"next_operation"`
	SourceDigest          string      `json:"source_digest"`
	CandidateSourceDigest string      `json:"candidate_source_digest"`
	RevisionCandidateID   string      `json:"revision_candidate_id"`
	ComparisonDigest      string      `json:"comparison_digest"`
	ExecutionAllowed      bool        `json:"execution_allowed"`
	RepositoryWrites      int         `json:"repository_writes"`
}

func PrepareAcceptedRevisionNextRunStage(comparison AcceptedRevisionNextRunComparison, candidateSource []byte) (AcceptedRevisionNextRunStage, error) {
	if comparison.Schema != AcceptedRevisionNextRunComparisonSchema ||
		comparison.State != ReplayClosed ||
		comparison.Outcome != NextRunOutcomeImproved ||
		comparison.Reason != "SOURCE_REVISION_IMPROVED_ON_NEXT_RUN" ||
		comparison.ExecutionAllowed ||
		comparison.RepositoryWrites != 0 ||
		comparison.SourceDigest == "" ||
		comparison.CandidateSourceDigest == "" ||
		comparison.RevisionCandidateID == "" ||
		comparison.Activity == "" ||
		comparison.InputDigest == "" ||
		comparison.BaselineFailureCode == "" ||
		comparison.BaselineExecutionDigest == "" ||
		comparison.AcceptedExecutionDigest == "" ||
		comparison.NextCandidateExecutionDigest == "" ||
		len(comparison.BlockedBy) != 0 {
		return AcceptedRevisionNextRunStage{}, fmt.Errorf("accepted next-run comparison is not a closed improved result")
	}

	if got := digestBytes(candidateSource); got != comparison.CandidateSourceDigest {
		return AcceptedRevisionNextRunStage{}, fmt.Errorf("candidate source digest mismatch: got %s, want %s", got, comparison.CandidateSourceDigest)
	}

	canonical, err := json.Marshal(comparison)
	if err != nil {
		return AcceptedRevisionNextRunStage{}, fmt.Errorf("marshal comparison: %w", err)
	}

	return AcceptedRevisionNextRunStage{
		Schema:                AcceptedRevisionNextRunStageSchema,
		Decision:              AcceptedRevisionNextRunStageDecision,
		State:                 comparison.State,
		Outcome:               comparison.Outcome,
		Reason:                AcceptedRevisionNextRunStageReason,
		NextOperation:         AcceptedRevisionNextRunStageNextOp,
		SourceDigest:          comparison.SourceDigest,
		CandidateSourceDigest: comparison.CandidateSourceDigest,
		RevisionCandidateID:   comparison.RevisionCandidateID,
		ComparisonDigest:      digestBytes(canonical),
		ExecutionAllowed:      false,
		RepositoryWrites:      0,
	}, nil
}
