package valueexecution

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
)

const SelfImprovementSecurityBoundaryObservationSchema = "gooo.self-improvement.security-boundary.v1"

const (
	SelfImprovementSecurityBoundaryBound    = "BOUND"
	SelfImprovementSecurityBoundaryRejected = "REJECTED"
	SelfImprovementSecurityBoundaryUnknown  = "UNKNOWN"
)

type SelfImprovementSecurityBoundaryInput struct {
	WorkloadID           string `json:"workloadId"`
	Audience             string `json:"audience"`
	Freshness            string `json:"freshness"`
	AttestationDigest    string `json:"attestationDigest"`
	SVIDDigest           string `json:"svidDigest"`
	TrustBundleDigest    string `json:"trustBundleDigest"`
	WorkloadIdentityDigest string `json:"workloadIdentityDigest"`
}

type SelfImprovementSecurityBoundaryObservation struct {
	Schema                 string `json:"schema"`
	WorkloadID             string `json:"workloadId"`
	Audience               string `json:"audience"`
	Freshness              string `json:"freshness"`
	Status                 string `json:"status"`
	Reason                 string `json:"reason"`
	NonAuthorizing         bool   `json:"nonAuthorizing"`
	CryptographicVerified  bool   `json:"cryptographicVerified"`
	Digest                 string `json:"digest"`
}

func ObserveSelfImprovementSecurityBoundary(
	input SelfImprovementSecurityBoundaryInput,
) SelfImprovementSecurityBoundaryObservation {
	observation := SelfImprovementSecurityBoundaryObservation{
		Schema:                SelfImprovementSecurityBoundaryObservationSchema,
		WorkloadID:            input.WorkloadID,
		Audience:              input.Audience,
		Freshness:             input.Freshness,
		NonAuthorizing:        true,
		CryptographicVerified: false,
	}

	switch {
	case !strings.HasPrefix(input.WorkloadID, "spiffe://"):
		observation.Status = SelfImprovementSecurityBoundaryUnknown
		observation.Reason = "WORKLOAD_ID_INVALID"
	case strings.TrimSpace(input.Audience) == "":
		observation.Status = SelfImprovementSecurityBoundaryUnknown
		observation.Reason = "AUDIENCE_MISSING"
	case input.Freshness == "EXPIRED":
		observation.Status = SelfImprovementSecurityBoundaryRejected
		observation.Reason = "WORKLOAD_IDENTITY_EXPIRED"
	case input.Freshness != "FRESH":
		observation.Status = SelfImprovementSecurityBoundaryUnknown
		observation.Reason = "WORKLOAD_IDENTITY_FRESHNESS_UNKNOWN"
	case !isSelfImprovementSHA256(input.AttestationDigest) ||
		!isSelfImprovementSHA256(input.SVIDDigest) ||
		!isSelfImprovementSHA256(input.TrustBundleDigest) ||
		!isSelfImprovementSHA256(input.WorkloadIdentityDigest):
		observation.Status = SelfImprovementSecurityBoundaryUnknown
		observation.Reason = "SECURITY_DIGEST_INVALID"
	default:
		observation.Status = SelfImprovementSecurityBoundaryBound
		observation.Reason = "WORKLOAD_IDENTITY_BOUND"
	}

	observation.Digest = digestSelfImprovementSecurityBoundary(observation)
	return observation
}

func isSelfImprovementSHA256(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == sha256.Size
}

func digestSelfImprovementSecurityBoundary(
	observation SelfImprovementSecurityBoundaryObservation,
) string {
	payload, err := json.Marshal(observation)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}