package domaincapability

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

const EvidenceEnvelopeSchema = "meta-ontology.domain-capability.v1"

type EvidenceEnvelope struct {
	Schema                    string   `json:"schema"`
	Decision                  Decision `json:"decision"`
	ExpectedCapabilityCount  int      `json:"expected_capability_count"`
	ObservedCapabilityCount  int      `json:"observed_capability_count"`
	UnresolvedCapabilityCount int      `json:"unresolved_capability_count"`
	ObservationCount          int      `json:"observation_count"`
	EvidenceBound             bool     `json:"evidence_bound"`
	MinimumObservationCount   int      `json:"minimum_observation_count"`
	MaximumUnresolvedRatio    float64  `json:"maximum_unresolved_ratio"`
	RequireEvidenceDigest     bool     `json:"require_evidence_digest"`
	SourceDigest              string   `json:"source_digest"`
	ContractDigest            string   `json:"contract_digest"`
	NonExecuting              bool     `json:"non_executing"`
	NonAuthorizing            bool     `json:"non_authorizing"`
}

func NewEvidenceEnvelope(policy Policy, measurement Measurement, decision Decision, sourceDigest, contractDigest string) EvidenceEnvelope {
	return EvidenceEnvelope{
		Schema:                     EvidenceEnvelopeSchema,
		Decision:                   decision,
		ExpectedCapabilityCount:   measurement.ExpectedCapabilityCount,
		ObservedCapabilityCount:   measurement.ObservedCapabilityCount,
		UnresolvedCapabilityCount:  measurement.UnresolvedCapabilityCount,
		ObservationCount:           measurement.ObservationCount,
		EvidenceBound:              measurement.EvidenceBound,
		MinimumObservationCount:    policy.MinimumObservationCount,
		MaximumUnresolvedRatio:     policy.MaximumUnresolvedRatio,
		RequireEvidenceDigest:      policy.RequireEvidenceDigest,
		SourceDigest:                sourceDigest,
		ContractDigest:              contractDigest,
		NonExecuting:                true,
		NonAuthorizing:              true,
	}
}

func MarshalEvidenceEnvelope(envelope EvidenceEnvelope) ([]byte, error) {
	return json.Marshal(envelope)
}

func EvidenceEnvelopeDigest(envelope EvidenceEnvelope) (string, error) {
	payload, err := MarshalEvidenceEnvelope(envelope)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}