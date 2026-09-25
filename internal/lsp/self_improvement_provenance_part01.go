package lsp

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

const SelfImprovementProvenanceSchema = "gooo/lsp-self-improvement-provenance/v1"

const (
	SelfImprovementProvenanceVisible = "VISIBLE"
	SelfImprovementProvenanceUnknown = "UNKNOWN"
)

// SelfImprovementProvenanceObservation is an LSP-facing projection of a
// detached execution observation. It is display evidence, not an edit or an
// authorization grant.
type SelfImprovementProvenanceObservation struct {
	Schema                     string                               `json:"schema"`
	URI                        string                               `json:"uri"`
	Version                    int                                  `json:"version"`
	OriginDigest               string                               `json:"origin_digest,omitempty"`
	EnvironmentDigest          string                               `json:"environment_digest,omitempty"`
	CandidateStatus            valueexecution.SelfImprovementStatus `json:"candidate_status"`
	CandidateDigest            string                               `json:"candidate_digest,omitempty"`
	ExecutionObservationDigest string                               `json:"execution_observation_digest,omitempty"`
	Decision                   string                               `json:"decision"`
	Reason                     string                               `json:"reason"`
	NonAuthorizing             bool                                 `json:"non_authorizing"`
	ObservationDigest          string                               `json:"observation_digest"`
}

// ObserveSelfImprovementProvenance keeps a complete, non-UNKNOWN result
// visible to an editor without making the LSP surface an execution control.
func ObserveSelfImprovementProvenance(uri string, version int, observation valueexecution.ExecutedSelfImprovementObservation) SelfImprovementProvenanceObservation {
	value := SelfImprovementProvenanceObservation{
		Schema:                     SelfImprovementProvenanceSchema,
		URI:                        strings.TrimSpace(uri),
		Version:                    version,
		CandidateStatus:            observation.Status,
		CandidateDigest:            observation.CandidateDigest,
		ExecutionObservationDigest: observation.Digest,
		Decision:                   SelfImprovementProvenanceUnknown,
		Reason:                     observation.Reason,
		NonAuthorizing:             true,
		OriginDigest:               observation.Candidate.AfterOriginDigest,
		EnvironmentDigest:          observation.Candidate.EnvironmentDigest}
	switch {
	case value.URI == "":
		value.Reason = "MISSING_DOCUMENT_URI"
	case value.Version < 0:
		value.Reason = "MISSING_DOCUMENT_VERSION"
	case observation.Status == valueexecution.SelfImprovementStatusUnknown:
		value.Reason = observation.Reason
	case !cache.Digest(observation.CandidateDigest).Known():
		value.Reason = "MISSING_CANDIDATE_DIGEST"
	case !cache.Digest(observation.Digest).Known():
		value.Reason = "MISSING_OBSERVATION_DIGEST"
	default:
		value.Decision = SelfImprovementProvenanceVisible
		value.Reason = "SELF_IMPROVEMENT_STATUS_BOUND"
	}
	return finalizeSelfImprovementProvenance(value)
}

// ValidateSelfImprovementProvenance verifies the immutable editor envelope.
func ValidateSelfImprovementProvenance(value SelfImprovementProvenanceObservation) error {
	if value.Schema != SelfImprovementProvenanceSchema || value.URI == "" || value.Version < 0 || !value.NonAuthorizing || value.Reason == "" {
		return errors.New("self-improvement provenance identity is invalid")
	}
	if value.Decision != SelfImprovementProvenanceVisible && value.Decision != SelfImprovementProvenanceUnknown {
		return errors.New("self-improvement provenance decision is invalid")
	}
	if value.Decision == SelfImprovementProvenanceVisible &&
		(!cache.Digest(value.CandidateDigest).Known() || !cache.Digest(value.ExecutionObservationDigest).Known()) {
		return errors.New("visible self-improvement provenance is incomplete")
	}
	if !cache.Digest(value.ObservationDigest).Known() || value.ObservationDigest != selfImprovementProvenanceDigest(value) {
		return errors.New("self-improvement provenance digest is invalid")
	}
	return nil
}

func finalizeSelfImprovementProvenance(value SelfImprovementProvenanceObservation) SelfImprovementProvenanceObservation {
	value.ObservationDigest = selfImprovementProvenanceDigest(value)
	return value
}

func selfImprovementProvenanceDigest(value SelfImprovementProvenanceObservation) string {
	value.ObservationDigest = ""
	payload, _ := json.Marshal(value)
	return cache.HashBytes(payload).String()
}
