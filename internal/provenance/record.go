// Package provenance stores source-backed evidence as a durable, append-only
// canonical JSONL ledger. It records facts; it does not infer authority from
// Git refs, credentials, protection settings, or business names.
package provenance

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

	// The following fields are input-only compatibility aliases for the
	// prototype that preceded this contract. They are normalized into the
	// fields above and are never emitted in canonical JSONL.
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

// VerifiedClaim asks the store to prove that a verified event exists for one
// stable semantic ID and the exact expected semantic/graph digests.
type VerifiedClaim struct {
	SemanticID     string
	SemanticDigest string
	GraphDigest    string
}

// ReadOptions controls freshness and claim checks. ExpectedSourceHash is kept
// as a source-compatible alias for ExpectedSourceDigest.
type ReadOptions struct {
	ExpectedSourceDigest string
	ExpectedSourceHash   string
	RequireFresh         bool
	Now                  time.Time
	RequiredVerified     []VerifiedClaim
}

// Snapshot is the physical, validated ledger view. Digest is over the exact
// canonical JSONL bytes in append order, so reordering changes the digest.
type Snapshot struct {
	Records []Evidence
	Digest  string
}

// ErrConflict identifies an event ID or digest conflict.
var ErrConflict = errors.New("provenance event conflict")

// ErrChainGap identifies a missing, reordered, or mismatched predecessor.
var ErrChainGap = errors.New("provenance chain gap")

// ErrClaimNotVerified identifies candidate/deferred evidence used as proof.
var ErrClaimNotVerified = errors.New("provenance claim is not verified")

// ErrStaleSource identifies evidence bound to a different source snapshot.
var ErrStaleSource = errors.New("provenance source is stale")

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
// manifest. Byte offsets are offsets in the JSONL file, not character counts.
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

type lineError struct {
	kind string
	err  error
}

func (e *lineError) Error() string { return e.err.Error() }

func (e *lineError) Unwrap() error { return e.err }

type wireEvidence struct {
	Schema         int               `json:"schema"`
	ID             string            `json:"id"`
	EventID        string            `json:"event_id"`
	SemanticID     string            `json:"semantic_id"`
	Producer       string            `json:"producer"`
	Kind           EvidenceKind      `json:"kind"`
	Status         EvidenceStatus    `json:"status"`
	SourceSpan     wireSourceSpan    `json:"source_span"`
	SourceDigest   string            `json:"source_digest"`
	SemanticDigest string            `json:"semantic_digest"`
	GraphDigest    string            `json:"graph_digest"`
	Sequence       uint64            `json:"sequence"`
	Predecessor    *DigestLink       `json:"predecessor"`
	Attributes     map[string]string `json:"attributes"`
	Freshness      Freshness         `json:"freshness"`
	Hash           string            `json:"hash"`

	// Legacy aliases are recognized, normalized, and not re-emitted.
	Type        string `json:"type"`
	Subject     string `json:"subject"`
	GeneratedBy string `json:"generated_by"`
}

type wireSourceSpan struct {
	URI   string   `json:"uri"`
	File  string   `json:"file"`
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type canonicalEvidence struct {
	Schema         int                `json:"schema"`
	ID             string             `json:"id"`
	SemanticID     string             `json:"semantic_id"`
	Producer       string             `json:"producer"`
	Kind           EvidenceKind       `json:"kind"`
	Status         EvidenceStatus     `json:"status"`
	SourceSpan     SourceSpan         `json:"source_span"`
	SourceDigest   string             `json:"source_digest"`
	SemanticDigest string             `json:"semantic_digest"`
	GraphDigest    string             `json:"graph_digest"`
	Sequence       uint64             `json:"sequence"`
	Predecessor    *DigestLink        `json:"predecessor,omitempty"`
	Attributes     map[string]string  `json:"attributes,omitempty"`
	Freshness      canonicalFreshness `json:"freshness"`
	Hash           string             `json:"hash,omitempty"`
}

type canonicalFreshness struct {
	ProducedAt string `json:"produced_at"`
	ValidUntil string `json:"valid_until,omitempty"`
}

func (e Evidence) canonical(includeHash bool) (canonicalEvidence, error) {
	if e.Hash == "" && includeHash {
		return canonicalEvidence{}, fmt.Errorf("event hash is required")
	}
	result := canonicalEvidence{
		Schema:         e.Schema,
		ID:             e.ID,
		SemanticID:     e.SemanticID,
		Producer:       e.Producer,
		Kind:           e.Kind,
		Status:         e.Status,
		SourceSpan:     e.SourceSpan,
		SourceDigest:   e.SourceDigest,
		SemanticDigest: e.SemanticDigest,
		GraphDigest:    e.GraphDigest,
		Sequence:       e.Sequence,
		Predecessor:    e.Predecessor,
		Attributes:     e.Attributes,
		Freshness: canonicalFreshness{
			ProducedAt: e.Freshness.ProducedAt,
			ValidUntil: e.Freshness.ValidUntil,
		},
	}
	if includeHash {
		result.Hash = e.Hash
	}
	return result, nil
}

func marshalEvidence(evidence Evidence, includeHash bool) ([]byte, error) {
	canonical, err := evidence.canonical(includeHash)
	if err != nil {
		return nil, err
	}
	return json.Marshal(canonical)
}

// CanonicalRecord returns one canonical JSON object without a trailing LF.
// It is useful to adapters and deterministic golden fixtures.
func CanonicalRecord(evidence Evidence) ([]byte, error) {
	normalized, err := normalizeEvidence(evidence)
	if err != nil {
		return nil, err
	}
	if normalized.Sequence == 0 {
		return nil, fmt.Errorf("event sequence is required when canonicalizing a ledger record")
	}
	normalized, err = finishRecord(normalized)
	if err != nil {
		return nil, err
	}
	return marshalEvidence(normalized, true)
}

// CanonicalJSONL returns deterministic LF-terminated canonical records. It
// does not sort records: chain order is semantic and must be preserved.
func CanonicalJSONL(records ...Evidence) ([]byte, error) {
	var output bytes.Buffer
	for _, record := range records {
		line, err := CanonicalRecord(record)
		if err != nil {
			return nil, err
		}
		output.Write(line)
		output.WriteByte('\n')
	}
	return output.Bytes(), nil
}

func decodeEvidence(raw []byte) (Evidence, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var wire wireEvidence
	if err := decoder.Decode(&wire); err != nil {
		return Evidence{}, &lineError{kind: "invalid-json", err: err}
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return Evidence{}, &lineError{kind: "multiple-values", err: fmt.Errorf("multiple JSON values")}
		}
		return Evidence{}, &lineError{kind: "invalid-json", err: err}
	}
	if wire.ID != "" && wire.EventID != "" && wire.ID != wire.EventID {
		return Evidence{}, &lineError{kind: "conflicting-id", err: fmt.Errorf("id and event_id differ")}
	}
	id := wire.ID
	if id == "" {
		id = wire.EventID
	}
	return Evidence{
		Schema:     wire.Schema,
		ID:         id,
		SemanticID: wire.SemanticID,
		Producer:   wire.Producer,
		Kind:       wire.Kind,
		Status:     wire.Status,
		SourceSpan: SourceSpan{
			URI: wire.SourceSpan.URI, File: wire.SourceSpan.File,
			Start: wire.SourceSpan.Start, End: wire.SourceSpan.End,
		},
		SourceDigest:   wire.SourceDigest,
		SemanticDigest: wire.SemanticDigest,
		GraphDigest:    wire.GraphDigest,
		Sequence:       wire.Sequence,
		Predecessor:    wire.Predecessor,
		Attributes:     wire.Attributes,
		Freshness:      wire.Freshness,
		Hash:           wire.Hash,
		Type:           wire.Type,
		Subject:        wire.Subject,
		GeneratedBy:    wire.GeneratedBy,
	}, nil
}

func normalizeEvidence(evidence Evidence) (Evidence, error) {
	if evidence.Schema == 0 {
		evidence.Schema = SchemaVersion
	}
	if evidence.Schema != SchemaVersion {
		return Evidence{}, fmt.Errorf("unsupported schema %d", evidence.Schema)
	}
	var err error
	if evidence.ID, err = normalizeIdentifier(evidence.ID, "id"); err != nil {
		return Evidence{}, err
	}
	if evidence.Producer == "" {
		evidence.Producer = evidence.GeneratedBy
	}
	if evidence.Producer, err = normalizeIdentifier(evidence.Producer, "producer"); err != nil {
		return Evidence{}, err
	}
	if evidence.SemanticID == "" {
		evidence.SemanticID = evidence.Subject
	}
	if evidence.SemanticID == "" {
		evidence.SemanticID = evidence.ID
	}
	if evidence.SemanticID, err = normalizeIdentifier(evidence.SemanticID, "semantic_id"); err != nil {
		return Evidence{}, err
	}
	if evidence.Kind == "" {
		evidence.Kind = EvidenceKind(evidence.Type)
	}
	kind, err := normalizeIdentifier(string(evidence.Kind), "kind")
	if err != nil {
		return Evidence{}, err
	}
	evidence.Kind = EvidenceKind(kind)
	if evidence.Status == "" {
		if value, ok := evidence.Attributes["status"]; ok {
			evidence.Status = EvidenceStatus(value)
		}
	}
	if evidence.Status == "" {
		return Evidence{}, fmt.Errorf("status is required")
	}
	attributeStatus, hasAttributeStatus := evidence.Attributes["status"]
	evidence.Status = normalizeStatus(evidence.Status)
	if hasAttributeStatus && normalizeStatus(EvidenceStatus(attributeStatus)) != evidence.Status {
		return Evidence{}, fmt.Errorf("attributes.status does not match status")
	}
	if !validStatus(evidence.Status) {
		return Evidence{}, fmt.Errorf("unsupported status %q", evidence.Status)
	}
	if evidence.SourceDigest == "" {
		evidence.SourceDigest = evidence.Freshness.SourceHash
	}
	evidence.SourceDigest, err = normalizeDigest(evidence.SourceDigest, "source_digest")
	if err != nil {
		return Evidence{}, err
	}
	if evidence.Freshness.SourceHash != "" {
		freshnessDigest, digestErr := normalizeDigest(evidence.Freshness.SourceHash, "freshness.source_hash")
		if digestErr != nil {
			return Evidence{}, digestErr
		}
		if freshnessDigest != evidence.SourceDigest {
			return Evidence{}, fmt.Errorf("freshness source_hash does not match source_digest")
		}
	}
	evidence.Freshness.SourceHash = evidence.SourceDigest
	evidence.SemanticDigest, err = normalizeDigest(evidence.SemanticDigest, "semantic_digest")
	if err != nil {
		return Evidence{}, err
	}
	evidence.GraphDigest, err = normalizeDigest(evidence.GraphDigest, "graph_digest")
	if err != nil {
		return Evidence{}, err
	}
	evidence.SourceSpan, err = normalizeSourceSpan(evidence.SourceSpan)
	if err != nil {
		return Evidence{}, err
	}
	evidence.Attributes, err = normalizeAttributes(evidence.Attributes)
	if err != nil {
		return Evidence{}, err
	}
	evidence.Freshness, err = normalizeFreshness(evidence.Freshness)
	if err != nil {
		return Evidence{}, err
	}
	if evidence.Predecessor != nil {
		link := *evidence.Predecessor
		link.ID, err = normalizeIdentifier(link.ID, "predecessor.id")
		if err != nil {
			return Evidence{}, err
		}
		link.Digest, err = normalizeDigest(link.Digest, "predecessor.digest")
		if err != nil {
			return Evidence{}, err
		}
		evidence.Predecessor = &link
	}
	return evidence, nil
}

func normalizeStatus(status EvidenceStatus) EvidenceStatus {
	switch strings.ToLower(strings.TrimSpace(string(status))) {
	case "passed", "pass":
		return StatusVerified
	case "not-run", "not_run":
		return StatusDeferred
	default:
		return EvidenceStatus(strings.ToLower(strings.TrimSpace(string(status))))
	}
}

func validStatus(status EvidenceStatus) bool {
	switch status {
	case StatusVerified, StatusCandidate, StatusDeferred, StatusFailed, StatusRejected:
		return true
	default:
		return false
	}
}

func normalizeIdentifier(value, field string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("%s is required", field)
	}
	if strings.ContainsAny(value, "\r\n") {
		return "", fmt.Errorf("%s must not contain line breaks", field)
	}
	return value, nil
}

func normalizeDigest(value, field string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if len(value) != sha256.Size*2 {
		return "", fmt.Errorf("%s must be a %d-character SHA-256 hex digest", field, sha256.Size*2)
	}
	if _, err := hex.DecodeString(value); err != nil {
		return "", fmt.Errorf("%s must be a SHA-256 hex digest: %w", field, err)
	}
	return value, nil
}

func normalizeSourceSpan(span SourceSpan) (SourceSpan, error) {
	if span.URI == "" {
		span.URI = span.File
	}
	if span.URI == "" {
		return SourceSpan{}, fmt.Errorf("source_span.uri is required")
	}
	var err error
	span.URI, err = normalizeIdentifier(span.URI, "source_span.uri")
	if err != nil {
		return SourceSpan{}, err
	}
	span.File = ""
	if err := normalizePosition(span.Start, "source_span.start"); err != nil {
		return SourceSpan{}, err
	}
	if err := normalizePosition(span.End, "source_span.end"); err != nil {
		return SourceSpan{}, err
	}
	if positionAfter(span.Start, span.End) {
		return SourceSpan{}, fmt.Errorf("source_span.end precedes source_span.start")
	}
	return span, nil
}

func normalizePosition(position Position, field string) error {
	if position.Offset < 0 || position.Line < 1 || position.Column < 1 {
		return fmt.Errorf("%s must have offset >= 0 and positive line/column", field)
	}
	return nil
}

func positionAfter(left, right Position) bool {
	if left.Offset != right.Offset {
		return left.Offset > right.Offset
	}
	if left.Line != right.Line {
		return left.Line > right.Line
	}
	return left.Column > right.Column
}

func normalizeAttributes(attributes map[string]string) (map[string]string, error) {
	if len(attributes) == 0 {
		return nil, nil
	}
	result := make(map[string]string, len(attributes))
	for key, value := range attributes {
		key = strings.TrimSpace(key)
		if key == "" || strings.ContainsAny(key, "\r\n") {
			return nil, fmt.Errorf("attribute keys must be non-empty and line-free")
		}
		if _, exists := result[key]; exists {
			return nil, fmt.Errorf("attribute key %q is duplicated after normalization", key)
		}
		if strings.ContainsAny(value, "\r\n") {
			return nil, fmt.Errorf("attribute %q must not contain line breaks", key)
		}
		result[key] = value
	}
	return result, nil
}

func normalizeFreshness(freshness Freshness) (Freshness, error) {
	var err error
	freshness.ProducedAt, err = normalizeTimestamp(freshness.ProducedAt, "freshness.produced_at")
	if err != nil {
		return Freshness{}, err
	}
	if freshness.ValidUntil != "" {
		freshness.ValidUntil, err = normalizeTimestamp(freshness.ValidUntil, "freshness.valid_until")
		if err != nil {
			return Freshness{}, err
		}
		produced, _ := time.Parse(time.RFC3339Nano, freshness.ProducedAt)
		validUntil, _ := time.Parse(time.RFC3339Nano, freshness.ValidUntil)
		if !validUntil.After(produced) {
			return Freshness{}, fmt.Errorf("freshness.valid_until must be after freshness.produced_at")
		}
	}
	return freshness, nil
}

func normalizeTimestamp(value, field string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("%s is required", field)
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return "", fmt.Errorf("%s: %w", field, err)
	}
	return parsed.UTC().Format(time.RFC3339Nano), nil
}

func digestBytes(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}
