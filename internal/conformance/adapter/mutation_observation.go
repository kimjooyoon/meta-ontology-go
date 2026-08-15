package adapter

import (
	"fmt"
	"sort"
	"strings"
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

func validateMutationAttempt(attempt MutationAttempt, paths ObserverPaths) error {
	if strings.TrimSpace(attempt.Operation) == "" || strings.TrimSpace(attempt.Outcome) == "" {
		return fmt.Errorf("mutation attempt operation and outcome are required")
	}
	canonical, err := canonicalObserverPath(attempt.Path)
	if err != nil || canonical != attempt.Path {
		return fmt.Errorf("mutation attempt path is not canonical")
	}
	if canonical != paths.SourcePath && canonical != paths.OutputPath &&
		!observerPathContains(paths.TempRoot, canonical) {
		return fmt.Errorf("mutation attempt path is outside observer paths")
	}
	return nil
}

func mutationAttemptLess(left, right MutationAttempt) bool {
	if left.Path != right.Path {
		return left.Path < right.Path
	}
	if left.Operation != right.Operation {
		return left.Operation < right.Operation
	}
	return left.Outcome < right.Outcome
}

// CaptureUnverifiedMutation stores advisory data that can never satisfy the oracle.
func (o *NoWriteObserver) CaptureUnverifiedMutation(evidence MutationEvidence) error {
	if evidence.Status == MutationEvidenceVerified {
		return oracleError(OracleNW003, "public mutation capture cannot verify evidence")
	}
	if evidence.Status != MutationEvidenceMissing && evidence.Status != MutationEvidenceUnverified {
		return oracleError(OracleNW003, "unsupported mutation evidence status")
	}
	return o.captureMutation(evidence)
}

func (o *NoWriteObserver) captureVerifiedMutation(evidence verifiedMutationEvidence) error {
	return o.captureMutation(evidence.evidence)
}

func (o *NoWriteObserver) captureMutation(evidence MutationEvidence) error {
	if o == nil || o.stamp == nil || o.finished {
		return oracleError(OracleNW003, "observer mutation capture is closed")
	}
	if o.mutationCaptured {
		return oracleError(OracleNW003, "observer mutation capture is immutable")
	}
	o.mutation = evidence.clone()
	o.mutationCaptured = true
	return nil
}

func validateObserverMutation(observation NoWriteObservation) error {
	switch observation.Mutation.Status {
	case MutationEvidenceMissing:
		return oracleError(OracleNW001, "observer mutation-attempt evidence is missing")
	case MutationEvidenceUnverified:
		return oracleError(OracleNW003, "observer mutation-attempt evidence is not independently verified")
	case MutationEvidenceVerified:
		if observation.Mutation.Binding != observation.Binding {
			return oracleError(OracleID001, "mutation evidence binding does not match observer")
		}
		if err := validateVerifiedMutation(observation.Mutation, observation.Paths); err != nil {
			return oracleError(OracleNW003, "mutation evidence: "+err.Error())
		}
		if len(observation.Mutation.Attempts) != 0 {
			return oracleError(OracleNW004, "observer recorded an attempted mutation")
		}
		return nil
	default:
		return oracleError(OracleNW003, "observer mutation evidence status is invalid")
	}
}
