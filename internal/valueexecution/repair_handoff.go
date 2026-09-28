package valueexecution

import "errors"

type RepairHandoff struct {
	Schema           string      `json:"schema"`
	CandidateID      string      `json:"candidate_id"`
	CandidateDigest  string      `json:"candidate_digest"`
	ComparisonDigest string      `json:"comparison_digest"`
	TriggerState     ReplayState `json:"trigger_state"`
	TriggerReason    string      `json:"trigger_reason"`
	NextOperation    string      `json:"next_operation"`
	BlockedBy        []string    `json:"blocked_by"`
	Status           string      `json:"status"`
	ExecutionAllowed bool        `json:"execution_allowed"`
	RepositoryWrites int         `json:"repository_writes"`
}

const RepairHandoffSchema = "gooo/value-execution-repair-handoff/v1"
const RepairHandoffDeferred = "DEFERRED"

// ConsumeRepairCandidate validates a detached candidate and turns it into an
// explicit next-operation handoff. Consumption never grants execution or
// repository authority.
func ConsumeRepairCandidate(candidate RepairCandidate) (RepairHandoff, error) {
	if err := ValidateRepairCandidate(candidate); err != nil {
		return RepairHandoff{}, err
	}
	handoff := RepairHandoff{
		Schema:           RepairHandoffSchema,
		CandidateID:      candidate.CandidateID,
		CandidateDigest:  DigestRepairCandidate(candidate),
		ComparisonDigest: candidate.ComparisonDigest,
		TriggerState:     candidate.TriggerState,
		TriggerReason:    candidate.TriggerReason,
		NextOperation:    candidate.NextOperation,
		BlockedBy:        append([]string(nil), candidate.BlockedBy...),
		Status:           RepairHandoffDeferred,
		ExecutionAllowed: false,
		RepositoryWrites: 0,
	}
	if err := ValidateRepairHandoff(handoff); err != nil {
		return RepairHandoff{}, err
	}
	return handoff, nil
}

func DigestRepairCandidate(candidate RepairCandidate) string {
	return digestValue(candidate)
}

func ValidateRepairHandoff(handoff RepairHandoff) error {
	if handoff.Schema != RepairHandoffSchema || handoff.TriggerState != ReplayRefuted || handoff.Status != RepairHandoffDeferred {
		return errors.New("repair handoff identity is invalid")
	}
	if handoff.CandidateID == "" || !validDigest(handoff.CandidateDigest) || !validDigest(handoff.ComparisonDigest) {
		return errors.New("repair handoff identity is incomplete")
	}
	if handoff.TriggerReason == "" || handoff.NextOperation == "" || handoff.ExecutionAllowed || handoff.RepositoryWrites != 0 {
		return errors.New("repair handoff authority boundary is invalid")
	}
	seen := make(map[string]struct{}, len(handoff.BlockedBy))
	for _, item := range handoff.BlockedBy {
		if item == "" {
			return errors.New("repair handoff blocked frontier contains an empty item")
		}
		if _, exists := seen[item]; exists {
			return errors.New("repair handoff blocked frontier contains a duplicate item")
		}
		seen[item] = struct{}{}
	}
	return nil
}
