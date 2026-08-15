// Package provenance stores source-backed evidence as a durable, append-only
// canonical JSONL ledger. It records facts; it does not infer authority from
// Git refs, credentials, protection settings, or business names.
package provenance

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// SchemaVersion is the wire schema used by the ledger.
const SchemaVersion = 1

// EvidenceKind classifies the fact without granting it authority.
type EvidenceKind string

const (
	KindCompilerRun  EvidenceKind = "compiler-run"
	KindVerification EvidenceKind = "verification"
	KindComparison   EvidenceKind = "comparison"
	KindObservation  EvidenceKind = "observation"
)

// EvidenceStatus is deliberately separate from EvidenceKind. Candidate and
// deferred observations can be stored, but neither can satisfy a verified
// claim.
type EvidenceStatus string

const (
	StatusVerified  EvidenceStatus = "verified"
	StatusCandidate EvidenceStatus = "candidate"
	StatusDeferred  EvidenceStatus = "deferred"
	StatusFailed    EvidenceStatus = "failed"
	StatusRejected  EvidenceStatus = "rejected"
)

// Position identifies a source location. Offsets are zero-based; lines and
// columns are one-based, matching the compiler's source-span convention.
type Position struct {
	Offset int `json:"offset"`
	Line   int `json:"line"`
	Column int `json:"column"`
}

// SourceSpan binds an observation to the source that produced it.
type SourceSpan struct {
	URI   string   `json:"uri"`
	File  string   `json:"-"`
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// DigestLink is one immutable predecessor reference in the ledger chain.
type DigestLink struct {
	ID     string `json:"id"`
	Digest string `json:"digest"`
}

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

// Snapshot is the physical, validated ledger view. Digest covers exact
// canonical JSONL bytes in append order, so reordering changes the digest.
type Snapshot struct {
	Records []Evidence
	Digest  string
}

var (
	// ErrConflict identifies an event ID or digest conflict.
	ErrConflict = errors.New("provenance event conflict")
	// ErrChainGap identifies a missing, reordered, or mismatched predecessor.
	ErrChainGap = errors.New("provenance chain gap")
	// ErrClaimNotVerified identifies candidate/deferred evidence used as proof.
	ErrClaimNotVerified = errors.New("provenance claim is not verified")
	// ErrStaleSource identifies evidence bound to a different source snapshot.
	ErrStaleSource = errors.New("provenance source is stale")
)

// ConflictError reports a differing canonical event with an existing ID.
type ConflictError struct {
	ID     string
	Field  string
	Detail string
}

func (e *ConflictError) Error() string {
	if e.Field == "" {
		return fmt.Sprintf("event %q conflicts with existing canonical bytes: %s", e.ID, e.Detail)
	}
	return fmt.Sprintf("event %q conflicts in %s: %s", e.ID, e.Field, e.Detail)
}

func (e *ConflictError) Unwrap() error { return ErrConflict }

// ChainError reports a predecessor or sequence violation.
type ChainError struct {
	ID       string
	Expected string
	Actual   string
	Detail   string
}

func (e *ChainError) Error() string {
	return fmt.Sprintf("event %q has a provenance chain gap: %s (expected %q, got %q)", e.ID, e.Detail, e.Expected, e.Actual)
}

func (e *ChainError) Unwrap() error { return ErrChainGap }

// ClaimError reports why a verification claim cannot be satisfied.
type ClaimError struct {
	SemanticID string
	Kind       string
	Detail     string
}

func (e *ClaimError) Error() string {
	return fmt.Sprintf("semantic claim %q is not verified (%s): %s", e.SemanticID, e.Kind, e.Detail)
}

func (e *ClaimError) Unwrap() error { return ErrClaimNotVerified }

// CorruptionError identifies a malformed or integrity-invalid ledger line or
// manifest. Byte offsets are offsets in the JSONL file.
type CorruptionError struct {
	Path   string
	Line   int
	Offset int64
	Kind   string
	Detail string
	cause  error
}

func (e *CorruptionError) Error() string {
	return fmt.Sprintf("provenance corruption at %s:%d (byte %d, %s): %s", e.Path, e.Line, e.Offset, e.Kind, e.Detail)
}

func (e *CorruptionError) Unwrap() error { return e.cause }

// FreshnessError reports a valid event that does not match the requested
// source snapshot or freshness window.
type FreshnessError struct {
	ID       string
	Kind     string
	Expected string
	Actual   string
}

func (e *FreshnessError) Error() string {
	if e.Kind == "source-mismatch" {
		return fmt.Sprintf("event %q is stale: source digest %q, expected %q", e.ID, e.Actual, e.Expected)
	}
	return fmt.Sprintf("event %q is stale: %s", e.ID, e.Kind)
}

func (e *FreshnessError) Unwrap() error { return ErrStaleSource }
