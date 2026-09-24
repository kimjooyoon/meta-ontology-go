package valueexecution

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

const SelfImprovementReverseClosureObservationSchema = "gooo.self-improvement.reverse-closure.v1"

const (
	SelfImprovementReverseClosureCompleted   = "COMPLETED"
	SelfImprovementReverseClosureNotApplied  = "NOT_APPLIED"
	SelfImprovementReverseClosureUnknown     = "UNKNOWN"
)

type SelfImprovementReverseClosureInput struct {
	ForwardDigest          string
	Outcome                string
	OutcomeDigest          string
	SecurityBoundary      SelfImprovementExecutionSecurityBoundaryObservation
}

type SelfImprovementReverseClosureObservation struct {
	Schema                string `json:"schema"`
	ForwardDigest         string `json:"forwardDigest"`
	Outcome                string `json:"outcome"`
	OutcomeDigest         string `json:"outcomeDigest"`
	SecurityBoundaryStatus string `json:"securityBoundaryStatus"`
	SecurityBoundaryDigest string `json:"securityBoundaryDigest"`
	Status                string `json:"status"`
	Reason                string `json:"reason"`
	ReverseObservation    bool   `json:"reverseObservation"`
	NonAuthorizing        bool   `json:"nonAuthorizing"`
	Digest                string `json:"digest"`
}

func ObserveSelfImprovementReverseClosure(
	input SelfImprovementReverseClosureInput,
) SelfImprovementReverseClosureObservation {
	observation := SelfImprovementReverseClosureObservation{
		Schema:                 SelfImprovementReverseClosureObservationSchema,
		ForwardDigest:          input.ForwardDigest,
		Outcome:                input.Outcome,
		OutcomeDigest:          input.OutcomeDigest,
		SecurityBoundaryStatus: input.SecurityBoundary.Status,
		SecurityBoundaryDigest: input.SecurityBoundary.Digest,
		ReverseObservation:     true,
		NonAuthorizing:         true,
	}

	switch {
	case input.ForwardDigest == "" || input.OutcomeDigest == "":
		observation.Status = SelfImprovementReverseClosureUnknown
		observation.Reason = "REVERSE_EVIDENCE_MISSING"
	case input.SecurityBoundary.Status == SelfImprovementSecurityBoundaryRejected:
		observation.Status = SelfImprovementReverseClosureNotApplied
		observation.Reason = "SECURITY_BOUNDARY_REJECTED"
	case input.SecurityBoundary.Status != SelfImprovementSecurityBoundaryBound:
		observation.Status = SelfImprovementReverseClosureUnknown
		observation.Reason = "SECURITY_BOUNDARY_UNKNOWN"
	case input.Outcome == "REJECTED":
		observation.Status = SelfImprovementReverseClosureNotApplied
		observation.Reason = "EXECUTION_OUTCOME_REJECTED"
	case input.Outcome != "COMPLETED":
		observation.Status = SelfImprovementReverseClosureUnknown
		observation.Reason = "EXECUTION_OUTCOME_UNKNOWN"
	default:
		observation.Status = SelfImprovementReverseClosureCompleted
		observation.Reason = "REVERSE_CLOSURE_COMPLETED"
	}

	observation.Digest = digestSelfImprovementReverseClosure(observation)
	return observation
}

func digestSelfImprovementReverseClosure(
	observation SelfImprovementReverseClosureObservation,
) string {
	payload, err := json.Marshal(observation)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}