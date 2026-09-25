package provenance

import (
	"strings"
	"time"
)

// Evidence is one append-only event. ID is the event's stable identity and
// SemanticID is the stable semantic subject it describes. Hash is the
// SHA-256 of every canonical field except Hash itself, including timestamps.
// Therefore a timestamp can never replace content identity.
type Evidence struct {
	Schema         int               `json:"schema"`
	ID             string            `json:"id"`
	SemanticID     string            `json:"semantic_id"`
	Producer       string            `json:"producer"`
	Kind           EvidenceKind      `json:"kind"`
	Status         EvidenceStatus    `json:"status"`
	SourceSpan     SourceSpan        `json:"source_span"`
	SourceDigest   string            `json:"source_digest"`
	SemanticDigest string            `json:"semantic_digest"`
	GraphDigest    string            `json:"graph_digest"`
	Sequence       uint64            `json:"sequence"`
	Predecessor    *DigestLink       `json:"predecessor,omitempty"`
	Attributes     map[string]string `json:"attributes,omitempty"`
	Freshness      Freshness         `json:"freshness"`
	Hash           string            `json:"hash"`

	// Input-only compatibility aliases from the superseded prototype. They
	// are normalized into the fields above and never emitted in JSONL.
	Type        string `json:"-"`
	Subject     string `json:"-"`
	GeneratedBy string `json:"-"`
}

// Freshness carries time-window metadata. SourceHash is accepted as a legacy
// alias for SourceDigest and is always normalized to the same digest.
type Freshness struct {
	SourceHash string `json:"source_hash,omitempty"`
	ProducedAt string `json:"produced_at"`
	ValidUntil string `json:"valid_until,omitempty"`
}

// NewFreshness creates deterministic timestamp metadata for an event.
func NewFreshness(sourceDigest string, producedAt, validUntil time.Time) Freshness {
	result := Freshness{SourceHash: strings.ToLower(strings.TrimSpace(sourceDigest))}
	if !producedAt.IsZero() {
		result.ProducedAt = producedAt.UTC().Format(time.RFC3339Nano)
	}
	if !validUntil.IsZero() {
		result.ValidUntil = validUntil.UTC().Format(time.RFC3339Nano)
	}
	return result
}

// VerifiedClaim asks the store to prove one stable semantic ID and exact
// semantic/graph digests were emitted with verified status.
type VerifiedClaim struct {
	SemanticID     string
	SemanticDigest string
	GraphDigest    string
}

// ReadOptions controls freshness and claim checks. ExpectedSourceHash is a
// source-compatible alias for ExpectedSourceDigest.
type ReadOptions struct {
	ExpectedSourceDigest string
	ExpectedSourceHash   string
	RequireFresh         bool
	Now                  time.Time
	RequiredVerified     []VerifiedClaim
}
