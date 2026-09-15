package adapter

import (
	"bytes"
	"fmt"
)

// CompareEvidence compares neutral payloads, never producer labels.
func CompareEvidence(left, right EvidenceArtifact) error {
	leftPayload, err := left.CanonicalPayload()
	if err != nil {
		return fmt.Errorf("left evidence: %w", err)
	}
	rightPayload, err := right.CanonicalPayload()
	if err != nil {
		return fmt.Errorf("right evidence: %w", err)
	}
	if bytes.Equal(leftPayload, rightPayload) {
		return nil
	}
	return fmt.Errorf("evidence mismatch: %s != %s", digest(leftPayload), digest(rightPayload))
}

// ProjectEvidence projects a validated response into the shared evidence contract.
func (r Response) ProjectEvidence(producer string, stage Stage) (EvidenceArtifact, error) {
	normalized, err := r.Normalized()
	if err != nil {
		return EvidenceArtifact{}, err
	}
	facts := []EvidenceFact{
		{ID: normalized.Fixture + "/operation", Kind: "operation", Value: string(normalized.Operation)},
		{ID: normalized.Fixture + "/status", Kind: "status", Value: string(normalized.Status)},
		{ID: normalized.Fixture + "/promotion", Kind: "promotion", Value: fmt.Sprintf("%t", normalized.PromotionEligible)},
	}
	if normalized.Failure != nil {
		facts = append(facts, EvidenceFact{
			ID:    normalized.Fixture + "/failure",
			Kind:  "failure",
			Value: normalized.Failure.Code,
		})
	}
	if normalized.Observed.SemanticDigest != "" {
		facts = append(facts, EvidenceFact{
			ID:    normalized.Fixture + "/semantic",
			Kind:  "semantic-digest",
			Value: normalized.Observed.SemanticDigest,
		})
	}
	artifact := EvidenceArtifact{
		Producer: producer,
		Bundle: EvidenceBundle{
			Schema:   EvidenceSchema,
			Stage:    stage,
			Fixture:  normalized.Fixture,
			Decision: string(normalized.Status),
			Facts:    facts,
		},
	}
	if err := artifact.Validate(); err != nil {
		return EvidenceArtifact{}, err
	}
	return artifact, nil
}
