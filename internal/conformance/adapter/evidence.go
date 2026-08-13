package adapter

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

// Stage identifies the trust boundary that produced evidence.
type Stage uint8

const (
	StageGoBaseline Stage = iota
	StageDualEvidence
	StageGoooFallback
	StageGoooAuthoritative
)

// EvidenceFact is a stable-ID assertion projected from a response.
type EvidenceFact struct {
	ID    string `json:"id"`
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

// EvidenceBundle is independent of producer identity and is comparison input.
type EvidenceBundle struct {
	Schema   string         `json:"schema"`
	Stage    Stage          `json:"stage"`
	Fixture  string         `json:"fixture"`
	Decision string         `json:"decision"`
	Facts    []EvidenceFact `json:"facts"`
}

// EvidenceArtifact identifies the producer while retaining a neutral bundle.
type EvidenceArtifact struct {
	Producer string         `json:"producer"`
	Bundle   EvidenceBundle `json:"bundle"`
}

// EvidenceManifest is an append-only, independently checkable digest record.
type EvidenceManifest struct {
	Schema        string `json:"schema"`
	Producer      string `json:"producer"`
	Stage         Stage  `json:"stage"`
	Fixture       string `json:"fixture"`
	Decision      string `json:"decision"`
	PayloadSHA256 string `json:"payload_sha256"`
}

// Validate checks only protocol invariants; it never promotes a result.
func (a EvidenceArtifact) Validate() error {
	if a.Producer != "go" && a.Producer != "gooo" {
		return fmt.Errorf("unsupported evidence producer %q", a.Producer)
	}
	bundle := a.Bundle
	if bundle.Schema != EvidenceSchema {
		return fmt.Errorf("unsupported evidence schema %q", bundle.Schema)
	}
	if bundle.Stage > StageGoooAuthoritative {
		return fmt.Errorf("unsupported evidence stage %d", bundle.Stage)
	}
	if strings.TrimSpace(bundle.Fixture) == "" || strings.TrimSpace(bundle.Decision) == "" {
		return fmt.Errorf("evidence fixture and decision are required")
	}
	for _, fact := range bundle.Facts {
		if strings.TrimSpace(fact.ID) == "" || strings.TrimSpace(fact.Kind) == "" {
			return fmt.Errorf("evidence facts require stable id and kind")
		}
	}
	return nil
}

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
