package provenance

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
)

const OriginEnvironmentSchema = "gooo/origin-environment-observation/v1"

type OriginEnvironmentStatus string

const (
	OriginEnvironmentStatusUnknown OriginEnvironmentStatus = "UNKNOWN"
	OriginEnvironmentStatusBound   OriginEnvironmentStatus = "BOUND"
)

// OriginEnvironmentObservation keeps declaration-to-metric origin separate
// from the environment in which the observation was produced.
type OriginEnvironmentObservation struct {
	Schema            string                  `json:"schema"`
	OriginDigest      string                  `json:"origin_digest,omitempty"`
	EnvironmentDigest string                  `json:"environment_digest,omitempty"`
	Status            OriginEnvironmentStatus `json:"status"`
	Reason            string                  `json:"reason"`
	NonAuthorizing    bool                    `json:"non_authorizing"`
	EvidenceDigest    string                  `json:"evidence_digest"`
}

// ObserveOriginEnvironment binds only complete origin evidence to a complete
// environment identity. It does not claim that either result is beneficial.
func ObserveOriginEnvironment(origin OriginChainObservation, environmentDigest string) OriginEnvironmentObservation {
	observation := OriginEnvironmentObservation{
		Schema:            OriginEnvironmentSchema,
		OriginDigest:      strings.TrimSpace(origin.Digest),
		EnvironmentDigest: strings.TrimSpace(environmentDigest),
		Status:            OriginEnvironmentStatusUnknown,
		Reason:            "ORIGIN_ENVIRONMENT_INCOMPLETE",
		NonAuthorizing:    true,
	}
	switch {
	case !origin.Comparable():
		observation.Reason = "ORIGIN_OBSERVATION_INCOMPLETE"
	case !validOriginEnvironmentDigest(observation.EnvironmentDigest):
		observation.Reason = "ENVIRONMENT_DIGEST_INVALID"
	default:
		observation.Status = OriginEnvironmentStatusBound
		observation.Reason = "ORIGIN_BOUND_TO_ENVIRONMENT"
	}
	observation.EvidenceDigest = originEnvironmentEvidenceDigest(observation)
	return observation
}

type OriginEnvironmentTransition string

const (
	OriginEnvironmentTransitionUnknown   OriginEnvironmentTransition = "UNKNOWN"
	OriginEnvironmentTransitionUnchanged OriginEnvironmentTransition = "UNCHANGED"
	OriginEnvironmentTransitionChanged   OriginEnvironmentTransition = "CHANGED"
)

// CompareOriginEnvironments never treats incomplete provenance as unchanged.
func CompareOriginEnvironments(before, after OriginEnvironmentObservation) OriginEnvironmentTransition {
	if before.Status != OriginEnvironmentStatusBound || after.Status != OriginEnvironmentStatusBound ||
		before.EvidenceDigest == "" || after.EvidenceDigest == "" {
		return OriginEnvironmentTransitionUnknown
	}
	if before.EvidenceDigest == after.EvidenceDigest {
		return OriginEnvironmentTransitionUnchanged
	}
	return OriginEnvironmentTransitionChanged
}

func validOriginEnvironmentDigest(value string) bool {
	if isSHA256Digest(value) {
		return true
	}
	if len(value) != sha256.Size*2 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func originEnvironmentEvidenceDigest(value OriginEnvironmentObservation) string {
	value.EvidenceDigest = ""
	payload, _ := json.Marshal(value)
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:])
}
