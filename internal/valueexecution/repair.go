package valueexecution

import "errors"

type RepairCandidate struct {
	Schema            string      `json:"schema"`
	CandidateID       string      `json:"candidate_id"`
	ComparisonDigest  string      `json:"comparison_digest"`
	TriggerState      ReplayState `json:"trigger_state"`
	TriggerReason     string      `json:"trigger_reason"`
	Target            string      `json:"target"`
	ExecutionAllowed  bool        `json:"execution_allowed"`
	RepositoryWrites  int         `json:"repository_writes"`
	NextOperation     string      `json:"next_operation"`
	BlockedBy         []string    `json:"blocked_by"`
}

const RepairCandidateSchema = "gooo/value-execution-repair-candidate/v1"

// ProposeRepair turns only a known replay contradiction into a non-executing
// candidate. It never edits source, runs a repair, or treats UNKNOWN as a
// repair trigger.
func ProposeRepair(comparison ReplayComparison) (RepairCandidate, error) {
	if comparison.Schema != ReplayComparisonSchema || comparison.State != ReplayRefuted || comparison.Reason == "" || comparison.NextOperation == "" {
		return RepairCandidate{}, errors.New("repair candidate requires a refuted replay comparison")
	}
	digest := DigestReplayComparison(comparison)
	return RepairCandidate{
		Schema:           RepairCandidateSchema,
		CandidateID:      "gooo://repair-candidate/" + digest[len("sha256:"):16],
		ComparisonDigest: digest,
		TriggerState:     comparison.State,
		TriggerReason:    comparison.Reason,
		Target:           "VALUE_EXECUTION_REPLAY",
		ExecutionAllowed: false,
		RepositoryWrites: 0,
		NextOperation:    comparison.NextOperation,
		BlockedBy:        append([]string(nil), comparison.BlockedBy...),
	}, nil
}

func DigestReplayComparison(comparison ReplayComparison) string {
	return digestValue(comparison)
}
