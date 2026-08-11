// Package bootstrap contains producer-independent evidence records for the
// Go-hosted bootstrap baseline. It has no dependency on compiler internals.
package bootstrap

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

const (
	SchemaVersion         = "gooo/bootstrap-evidence/v1"
	StageGoHostedBaseline = "go-hosted-baseline"
	ProducerGo            = "gooo://host/compiler/go"
	VerifierGo            = "gooo://host/verifier/go"
	StatusPass            = "pass"
	StatusFail            = "fail"
	StatusDeferred        = "deferred"
	StatusNotRun          = "not-run"
)

// Checks records the status of each baseline gate without hiding unavailable
// semantic or bootstrap checks behind a successful Go build.
type Checks struct {
	Format           string `json:"format"`
	Vet              string `json:"vet"`
	Test             string `json:"test"`
	Race             string `json:"race"`
	SemanticCLI      string `json:"semantic_cli"`
	BootstrapCompare string `json:"bootstrap_compare"`
}

// Evidence is a canonical, source-bound envelope for one Go-hosted run.
type Evidence struct {
	Schema            string  `json:"schema"`
	Stage             string  `json:"stage"`
	Producer          string  `json:"producer"`
	Verifier          string  `json:"verifier"`
	SourceDigest      *string `json:"source_digest"`
	SemanticDigest    *string `json:"semantic_digest"`
	ProvenanceDigest  *string `json:"provenance_digest"`
	Decision          string  `json:"decision"`
	EvidenceStatus    string  `json:"evidence_status"`
	PromotionEligible bool    `json:"promotion_eligible"`
	Checks            Checks  `json:"checks"`
}

// NewGoHostedBaseline records real source evidence while keeping semantic
// evidence deferred until the semantic CLI and provenance publisher exist.
func NewGoHostedBaseline(source []byte) Evidence {
	sourceDigest := DigestBytes(source)
	return Evidence{
		Schema:         SchemaVersion,
		Stage:          StageGoHostedBaseline,
		Producer:       ProducerGo,
		Verifier:       VerifierGo,
		SourceDigest:   &sourceDigest,
		Decision:       StatusDeferred,
		EvidenceStatus: StatusDeferred,
		Checks: Checks{
			Format: StatusPass, Vet: StatusPass, Test: StatusPass, Race: StatusPass,
			SemanticCLI: StatusDeferred, BootstrapCompare: StatusNotRun,
		},
	}
}

// Validate rejects missing evidence and any claim of promotion without the
// semantic and provenance records required by the bootstrap contract.
func (e Evidence) Validate() error {
	if e.Schema != SchemaVersion || e.Stage != StageGoHostedBaseline {
		return fmt.Errorf("unsupported Go-hosted evidence identity")
	}
	if e.Producer != ProducerGo || e.Verifier != VerifierGo {
		return fmt.Errorf("unexpected Go-hosted evidence producer or verifier")
	}
	if e.SourceDigest == nil || !validDigest(pointerValue(e.SourceDigest)) {
		return fmt.Errorf("source digest is required")
	}
	if e.SemanticDigest != nil && !validDigest(pointerValue(e.SemanticDigest)) {
		return fmt.Errorf("semantic digest is invalid")
	}
	if e.ProvenanceDigest != nil && !validDigest(pointerValue(e.ProvenanceDigest)) {
		return fmt.Errorf("provenance digest is invalid")
	}
	if !validStatus(e.Decision) || !validStatus(e.EvidenceStatus) {
		return fmt.Errorf("unsupported evidence decision or status")
	}
	if err := e.Checks.Validate(); err != nil {
		return err
	}
	if e.Checks.SemanticCLI != StatusPass && e.SemanticDigest != nil {
		return fmt.Errorf("semantic digest cannot precede a passing semantic CLI")
	}
	if e.PromotionEligible {
		if e.Decision != StatusPass || e.EvidenceStatus != StatusPass {
			return fmt.Errorf("only passing evidence can be promotion eligible")
		}
		if e.SemanticDigest == nil || e.ProvenanceDigest == nil {
			return fmt.Errorf("promotion requires semantic and provenance digests")
		}
	}
	if (e.Decision == StatusDeferred || e.Decision == StatusNotRun) && e.PromotionEligible {
		return fmt.Errorf("unavailable evidence cannot be promotion eligible")
	}
	return nil
}

// CanonicalJSON returns deterministic JSONL for an append-only evidence log.
func (e Evidence) CanonicalJSON() ([]byte, error) {
	if err := e.Validate(); err != nil {
		return nil, err
	}
	data, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

// Digest returns the identity of the validated evidence envelope.
func (e Evidence) Digest() (string, error) {
	data, err := e.CanonicalJSON()
	if err != nil {
		return "", err
	}
	return DigestBytes(data), nil
}

// DigestBytes returns a lowercase SHA-256 digest for content-addressed input.
func DigestBytes(content []byte) string {
	digest := sha256.Sum256(content)
	return hex.EncodeToString(digest[:])
}

func (c Checks) Validate() error {
	checks := []string{c.Format, c.Vet, c.Test, c.Race, c.SemanticCLI, c.BootstrapCompare}
	for _, status := range checks {
		if !validStatus(status) {
			return fmt.Errorf("unsupported check status %q", status)
		}
	}
	return nil
}

func pointerValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func validStatus(status string) bool {
	return status == StatusPass || status == StatusFail || status == StatusDeferred || status == StatusNotRun
}

func validDigest(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	for _, character := range value {
		if (character < '0' || character > '9') && (character < 'a' || character > 'f') {
			return false
		}
	}
	return true
}
