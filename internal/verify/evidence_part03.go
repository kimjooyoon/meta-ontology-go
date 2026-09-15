package verify

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// Manifest returns a digest over the canonical producer-independent payload.
func (artifact EvidenceArtifact) Manifest() (EvidenceManifest, error) {
	payload, err := artifact.CanonicalPayload()
	if err != nil {
		return EvidenceManifest{}, err
	}
	return EvidenceManifest{
		Schema:        artifact.Bundle.Schema,
		Producer:      artifact.Producer,
		Stage:         artifact.Bundle.Stage,
		Fixture:       artifact.Bundle.Fixture,
		Decision:      artifact.Bundle.Decision,
		PayloadSHA256: payloadDigest(payload),
	}, nil
}

// ManifestJSON returns deterministic JSONL for append-only evidence logs.
func (artifact EvidenceArtifact) ManifestJSON() ([]byte, error) {
	manifest, err := artifact.Manifest()
	if err != nil {
		return nil, err
	}
	encoded, err := json.Marshal(manifest)
	if err != nil {
		return nil, fmt.Errorf("marshal evidence manifest: %w", err)
	}
	return append(encoded, '\n'), nil
}

// CompareEvidence compares Go and gooo outputs without trusting producer names.
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
	return fmt.Errorf("evidence mismatch: %s != %s", payloadDigest(leftPayload), payloadDigest(rightPayload))
}
func payloadDigest(payload []byte) string {
	digest := sha256.Sum256(payload)
	return hex.EncodeToString(digest[:])
}
