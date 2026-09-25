package adapter

import (
	"fmt"
	"sort"
)

// MutationEvidenceStatus distinguishes observer-owned mutation traces from claims.
type MutationEvidenceStatus string

const (
	MutationEvidenceMissing    MutationEvidenceStatus = "missing"
	MutationEvidenceUnverified MutationEvidenceStatus = "unverified"
	MutationEvidenceVerified   MutationEvidenceStatus = "verified"
)

// MutationAttempt is a bound observer record of an attempted filesystem write.
type MutationAttempt struct {
	Path      string `json:"path"`
	Operation string `json:"operation"`
	Outcome   string `json:"outcome"`
}

// MutationEvidence records the observer's independent write-attempt trace.
type MutationEvidence struct {
	Status   MutationEvidenceStatus `json:"status"`
	Binding  ObservationBinding     `json:"binding"`
	Attempts []MutationAttempt      `json:"attempts,omitempty"`
}
type verifiedMutationEvidence struct {
	evidence MutationEvidence
}

func missingMutationEvidence() MutationEvidence {
	return MutationEvidence{Status: MutationEvidenceMissing}
}
func (e MutationEvidence) clone() MutationEvidence {
	e.Attempts = append([]MutationAttempt{}, e.Attempts...)
	return e
}
func newVerifiedMutationEvidence(
	evidence MutationEvidence, paths ObserverPaths,
) (verifiedMutationEvidence, error) {
	if evidence.Status != MutationEvidenceVerified {
		return verifiedMutationEvidence{}, fmt.Errorf("verified mutation status is required")
	}
	if err := validateVerifiedMutation(evidence, paths); err != nil {
		return verifiedMutationEvidence{}, err
	}
	return verifiedMutationEvidence{evidence: evidence.clone()}, nil
}
func validateVerifiedMutation(evidence MutationEvidence, paths ObserverPaths) error {
	if err := evidence.Binding.validate(); err != nil {
		return fmt.Errorf("mutation binding: %w", err)
	}
	if !sort.SliceIsSorted(evidence.Attempts, func(i, j int) bool {
		return mutationAttemptLess(evidence.Attempts[i], evidence.Attempts[j])
	}) {
		return fmt.Errorf("mutation attempts are not canonical")
	}
	for index, attempt := range evidence.Attempts {
		if err := validateMutationAttempt(attempt, paths); err != nil {
			return err
		}
		if index > 0 && evidence.Attempts[index-1] == attempt {
			return fmt.Errorf("mutation attempt is duplicated")
		}
	}
	return nil
}
