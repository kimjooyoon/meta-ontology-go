package valueexecution

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

const SelfImprovementExecutionSecurityBoundaryObservationSchema = "gooo.self-improvement.execution-security-boundary.v1"

const (
	SelfImprovementExecutionSecurityBoundaryReady    = "READY"
	SelfImprovementExecutionSecurityBoundaryRejected = "REJECTED"
	SelfImprovementExecutionSecurityBoundaryUnknown  = "UNKNOWN"
)

type SelfImprovementExecutionSecurityBoundaryInput struct {
	RequestID        string
	RequestDigest    string
	SecurityBoundary SelfImprovementSecurityBoundaryObservation
}

type SelfImprovementExecutionSecurityBoundaryObservation struct {
	Schema                  string `json:"schema"`
	RequestID               string `json:"requestId"`
	RequestDigest           string `json:"requestDigest"`
	SecurityBoundaryStatus  string `json:"securityBoundaryStatus"`
	SecurityBoundaryDigest  string `json:"securityBoundaryDigest"`
	Status                  string `json:"status"`
	Reason                  string `json:"reason"`
	NonAuthorizing          bool   `json:"nonAuthorizing"`
	Digest                  string `json:"digest"`
}

func ObserveSelfImprovementExecutionSecurityBoundary(
	input SelfImprovementExecutionSecurityBoundaryInput,
) SelfImprovementExecutionSecurityBoundaryObservation {
	observation := SelfImprovementExecutionSecurityBoundaryObservation{
		Schema:                 SelfImprovementExecutionSecurityBoundaryObservationSchema,
		RequestID:              input.RequestID,
		RequestDigest:          input.RequestDigest,
		SecurityBoundaryStatus: input.SecurityBoundary.Status,
		SecurityBoundaryDigest: input.SecurityBoundary.Digest,
		NonAuthorizing:         true,
	}

	switch {
	case input.RequestID == "" || input.RequestDigest == "":
		observation.Status = SelfImprovementExecutionSecurityBoundaryUnknown
		observation.Reason = "EXECUTION_REQUEST_IDENTITY_MISSING"
	case input.SecurityBoundary.Status == SelfImprovementSecurityBoundaryRejected:
		observation.Status = SelfImprovementExecutionSecurityBoundaryRejected
		observation.Reason = "SECURITY_BOUNDARY_REJECTED"
	case input.SecurityBoundary.Status != SelfImprovementSecurityBoundaryBound ||
		input.SecurityBoundary.Digest == "":
		observation.Status = SelfImprovementExecutionSecurityBoundaryUnknown
		observation.Reason = "SECURITY_BOUNDARY_UNKNOWN"
	default:
		observation.Status = SelfImprovementExecutionSecurityBoundaryReady
		observation.Reason = "EXECUTION_SECURITY_BOUNDARY_OBSERVED"
	}

	observation.Digest = digestSelfImprovementExecutionSecurityBoundary(observation)
	return observation
}

func digestSelfImprovementExecutionSecurityBoundary(
	observation SelfImprovementExecutionSecurityBoundaryObservation,
) string {
	payload, err := json.Marshal(observation)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}