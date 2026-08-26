package adapter

import (
	"fmt"
	"sort"
)

// CanonicalPayload returns producer-independent JSONL with sorted facts.
func (a EvidenceArtifact) CanonicalPayload() ([]byte, error) {
	normalized, err := a.normalized()
	if err != nil {
		return nil, err
	}
	return jsonLine(normalized.Bundle)
}
func (a EvidenceArtifact) normalized() (EvidenceArtifact, error) {
	if err := a.Validate(); err != nil {
		return EvidenceArtifact{}, err
	}
	a.Bundle.Facts = append([]EvidenceFact(nil), a.Bundle.Facts...)
	sort.Slice(a.Bundle.Facts, func(i, j int) bool {
		left, right := a.Bundle.Facts[i], a.Bundle.Facts[j]
		if left.ID != right.ID {
			return left.ID < right.ID
		}
		if left.Kind != right.Kind {
			return left.Kind < right.Kind
		}
		return left.Value < right.Value
	})
	for i := 1; i < len(a.Bundle.Facts); i++ {
		if a.Bundle.Facts[i].ID == a.Bundle.Facts[i-1].ID {
			return EvidenceArtifact{}, fmt.Errorf("duplicate evidence fact id %q", a.Bundle.Facts[i].ID)
		}
	}
	if a.Bundle.Facts == nil {
		a.Bundle.Facts = []EvidenceFact{}
	}
	return a, nil
}

// Manifest returns the digest over the producer-independent payload.
func (a EvidenceArtifact) Manifest() (EvidenceManifest, error) {
	payload, err := a.CanonicalPayload()
	if err != nil {
		return EvidenceManifest{}, err
	}
	return EvidenceManifest{
		Schema:        a.Bundle.Schema,
		Producer:      a.Producer,
		Stage:         a.Bundle.Stage,
		Fixture:       a.Bundle.Fixture,
		Decision:      a.Bundle.Decision,
		PayloadSHA256: digest(payload),
	}, nil
}

// ManifestJSON emits deterministic JSONL for an append-only evidence log.
func (a EvidenceArtifact) ManifestJSON() ([]byte, error) {
	manifest, err := a.Manifest()
	if err != nil {
		return nil, err
	}
	return jsonLine(manifest)
}
