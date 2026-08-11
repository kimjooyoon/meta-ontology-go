package verify

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

const (
	EvidenceSchemaVersion = "gooo/evidence/v1"
	EvidenceProducerGo    = "go"
	EvidenceProducerGooo  = "gooo"
)

// ConformanceStage names the staged verifier trust boundary.
type ConformanceStage uint8

const (
	StageGoBaseline ConformanceStage = iota
	StageDualEvidence
	StageGoooFallback
	StageGoooAuthoritative
)

// EvidenceFact is a canonical, stable-ID result from one verifier.
type EvidenceFact struct {
	ID    string `json:"id"`
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

// EvidenceBundle is the producer-independent payload compared by CI.
type EvidenceBundle struct {
	Schema   string           `json:"schema"`
	Stage    ConformanceStage `json:"stage"`
	Fixture  string           `json:"fixture"`
	Decision string           `json:"decision"`
	Facts    []EvidenceFact   `json:"facts"`
}

// EvidenceArtifact identifies the implementation that emitted a comparable
// bundle. Producer identity is excluded from the comparison payload.
type EvidenceArtifact struct {
	Producer string         `json:"producer"`
	Bundle   EvidenceBundle `json:"bundle"`
}

// EvidenceManifest records the independently verifiable payload digest.
type EvidenceManifest struct {
	Schema        string           `json:"schema"`
	Producer      string           `json:"producer"`
	Stage         ConformanceStage `json:"stage"`
	Fixture       string           `json:"fixture"`
	Decision      string           `json:"decision"`
	PayloadSHA256 string           `json:"payload_sha256"`
}

func (bundle EvidenceBundle) normalized() (EvidenceBundle, error) {
	if bundle.Schema != EvidenceSchemaVersion {
		return EvidenceBundle{}, fmt.Errorf("unsupported evidence schema %q", bundle.Schema)
	}
	if bundle.Stage > StageGoooAuthoritative {
		return EvidenceBundle{}, fmt.Errorf("unsupported conformance stage %d", bundle.Stage)
	}
	if strings.TrimSpace(bundle.Fixture) == "" || strings.TrimSpace(bundle.Decision) == "" {
		return EvidenceBundle{}, fmt.Errorf("evidence fixture and decision are required")
	}
	facts := append([]EvidenceFact(nil), bundle.Facts...)
	for _, fact := range facts {
		if strings.TrimSpace(fact.ID) == "" || strings.TrimSpace(fact.Kind) == "" {
			return EvidenceBundle{}, fmt.Errorf("evidence facts require stable id and kind")
		}
	}
	sort.Slice(facts, func(i, j int) bool {
		if facts[i].ID != facts[j].ID {
			return facts[i].ID < facts[j].ID
		}
		if facts[i].Kind != facts[j].Kind {
			return facts[i].Kind < facts[j].Kind
		}
		return facts[i].Value < facts[j].Value
	})
	for i := 1; i < len(facts); i++ {
		if facts[i].ID == facts[i-1].ID {
			return EvidenceBundle{}, fmt.Errorf("duplicate evidence fact id %q", facts[i].ID)
		}
	}
	if facts == nil {
		facts = []EvidenceFact{}
	}
	bundle.Facts = facts
	return bundle, nil
}

func validateEvidenceProducer(producer string) error {
	if producer != EvidenceProducerGo && producer != EvidenceProducerGooo {
		return fmt.Errorf("unsupported evidence producer %q", producer)
	}
	return nil
}

// CanonicalPayload returns producer-independent JSONL suitable for comparison.
func (artifact EvidenceArtifact) CanonicalPayload() ([]byte, error) {
	if err := validateEvidenceProducer(artifact.Producer); err != nil {
		return nil, err
	}
	bundle, err := artifact.Bundle.normalized()
	if err != nil {
		return nil, err
	}
	payload, err := json.Marshal(bundle)
	if err != nil {
		return nil, fmt.Errorf("marshal evidence bundle: %w", err)
	}
	return append(payload, '\n'), nil
}

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
