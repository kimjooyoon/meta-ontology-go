package valueexecution

import "errors"

type RepairCandidate struct {
	Schema           string      `json:"schema"`
	CandidateID      string      `json:"candidate_id"`
	ComparisonDigest string      `json:"comparison_digest"`
	TriggerState     ReplayState `json:"trigger_state"`
	TriggerReason    string      `json:"trigger_reason"`
	Target           string      `json:"target"`
	ExecutionAllowed bool        `json:"execution_allowed"`
	RepositoryWrites int         `json:"repository_writes"`
	NextOperation    string      `json:"next_operation"`
	BlockedBy        []string    `json:"blocked_by"`
}

const RepairCandidateSchema = "gooo/value-execution-repair-candidate/v1"

const repairCandidateDigestPrefix = "sha256:"
const repairCandidateDigestChars = 16

// ProposeRepair turns only a known replay contradiction into a non-executing
// candidate. It never edits source, runs a repair, or treats UNKNOWN as a
// repair trigger.
func ProposeRepair(comparison ReplayComparison) (RepairCandidate, error) {
	if comparison.Schema != ReplayComparisonSchema || comparison.State != ReplayRefuted || comparison.Reason == "" || comparison.NextOperation == "" {
		return RepairCandidate{}, errors.New("repair candidate requires a refuted replay comparison")
	}
	digest := DigestReplayComparison(comparison)
	candidate := RepairCandidate{
		Schema:           RepairCandidateSchema,
		CandidateID:      repairCandidateID(digest),
		ComparisonDigest: digest,
		TriggerState:     comparison.State,
		TriggerReason:    comparison.Reason,
		Target:           "VALUE_EXECUTION_REPLAY",
		ExecutionAllowed: false,
		RepositoryWrites: 0,
		NextOperation:    comparison.NextOperation,
		BlockedBy:        append([]string(nil), comparison.BlockedBy...),
	}
	if err := ValidateRepairCandidate(candidate); err != nil {
		return RepairCandidate{}, err
	}
	return candidate, nil
}

// ValidateRepairCandidate checks the detached candidate contract before a
// downstream consumer can treat it as a valid next operation. It grants no
// execution or repository authority.
func ValidateRepairCandidate(candidate RepairCandidate) error {
	if candidate.Schema != RepairCandidateSchema || candidate.TriggerState != ReplayRefuted || candidate.Target != "VALUE_EXECUTION_REPLAY" {
		return errors.New("repair candidate identity is invalid")
	}
	if !validDigest(candidate.ComparisonDigest) {
		return errors.New("repair candidate comparison digest is invalid")
	}
	wantID := repairCandidateID(candidate.ComparisonDigest)
	if candidate.CandidateID != wantID {
		return errors.New("repair candidate id does not match comparison digest")
	}
	if candidate.TriggerReason == "" || candidate.NextOperation == "" || candidate.ExecutionAllowed || candidate.RepositoryWrites != 0 {
		return errors.New("repair candidate authority boundary is invalid")
	}
	seen := make(map[string]struct{}, len(candidate.BlockedBy))
	for _, item := range candidate.BlockedBy {
		if item == "" {
			return errors.New("repair candidate blocked frontier contains an empty item")
		}
		if _, exists := seen[item]; exists {
			return errors.New("repair candidate blocked frontier contains a duplicate item")
		}
		seen[item] = struct{}{}
	}
	return nil
}

func DigestReplayComparison(comparison ReplayComparison) string {
	return digestValue(comparison)
}

func repairCandidateID(comparisonDigest string) string {
	start := len(repairCandidateDigestPrefix)
	return "gooo://repair-candidate/" + comparisonDigest[start:start+repairCandidateDigestChars]
}
