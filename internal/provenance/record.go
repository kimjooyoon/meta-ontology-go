// Package provenance stores append-only PROV evidence as canonical JSON Lines.
//
// The package is deliberately independent from the semantic compiler. Each line
// is a self-contained evidence entity with a content hash and freshness facts.
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

const schemaVersion = 1

// Evidence is one append-only PROV evidence entity.
type Evidence struct {
	Schema      int               `json:"schema"`
	ID          string            `json:"id"`
	Type        string            `json:"type"`
	Subject     string            `json:"subject"`
	GeneratedBy string            `json:"generated_by"`
	Attributes  map[string]string `json:"attributes,omitempty"`
	Freshness   Freshness         `json:"freshness"`
	Hash        string            `json:"hash"`
}

// Freshness records the source snapshot and the time window for an evidence
// entity. Times are RFC3339 timestamps normalized to UTC when stored.
type Freshness struct {
	SourceHash string `json:"source_hash"`
	ProducedAt string `json:"produced_at"`
	ValidUntil string `json:"valid_until,omitempty"`
}

// NewFreshness creates normalized timestamp metadata for an evidence entity.
func NewFreshness(sourceHash string, producedAt, validUntil time.Time) Freshness {
	metadata := Freshness{SourceHash: strings.ToLower(strings.TrimSpace(sourceHash))}
	if !producedAt.IsZero() {
		metadata.ProducedAt = producedAt.UTC().Format(time.RFC3339Nano)
	}
	if !validUntil.IsZero() {
		metadata.ValidUntil = validUntil.UTC().Format(time.RFC3339Nano)
	}
	return metadata
}

// ReadOptions controls optional freshness checks during Read.
type ReadOptions struct {
	ExpectedSourceHash string
	RequireFresh       bool
	Now                time.Time
}

// Snapshot is a stable, sorted view of the records in a store. Digest is the
// SHA-256 of canonical sorted JSON Lines, including the final newline per line.
type Snapshot struct {
	Records []Evidence
	Digest  string
}

// CorruptionError identifies a malformed or integrity-invalid JSONL line.
type CorruptionError struct {
	Path   string
	Line   int
	Offset int64
	Kind   string
	Detail string
}

// ErrDuplicateID is the stable identity-conflict sentinel for Append.
var ErrDuplicateID = errors.New("duplicate evidence id")

// DuplicateError identifies an append rejected by the store's unique-ID rule.
type DuplicateError struct {
	ID string
}

func (e *DuplicateError) Error() string {
	return fmt.Sprintf("evidence ID %q already exists", e.ID)
}

func (e *DuplicateError) Unwrap() error { return ErrDuplicateID }

func (e *CorruptionError) Error() string {
	return fmt.Sprintf("provenance corruption at %s:%d (byte %d, %s): %s", e.Path, e.Line, e.Offset, e.Kind, e.Detail)
}

// FreshnessError reports a valid record that does not match the requested
// source snapshot or freshness window.
type FreshnessError struct {
	ID       string
	Kind     string
	Expected string
	Actual   string
}

func (e *FreshnessError) Error() string {
	if e.Kind == "source-mismatch" {
		return fmt.Sprintf("evidence %q is stale: source hash %q, expected %q", e.ID, e.Actual, e.Expected)
	}
	return fmt.Sprintf("evidence %q is stale: %s", e.ID, e.Kind)
}

type lineError struct {
	kind string
	err  error
}

func (e *lineError) Error() string { return e.err.Error() }

func (e *lineError) Unwrap() error { return e.err }

func normalizeEvidence(evidence Evidence) (Evidence, error) {
	if evidence.Schema == 0 {
		evidence.Schema = schemaVersion
	}
	if evidence.Schema != schemaVersion {
		return Evidence{}, fmt.Errorf("unsupported schema %d", evidence.Schema)
	}
	var err error
	if evidence.ID, err = normalizeIdentifier(evidence.ID, "id"); err != nil {
		return Evidence{}, err
	}
	if evidence.Type, err = normalizeIdentifier(evidence.Type, "type"); err != nil {
		return Evidence{}, err
	}
	if evidence.Subject, err = normalizeIdentifier(evidence.Subject, "subject"); err != nil {
		return Evidence{}, err
	}
	if evidence.GeneratedBy, err = normalizeIdentifier(evidence.GeneratedBy, "generated_by"); err != nil {
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
	return evidence, nil
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

func normalizeAttributes(attributes map[string]string) (map[string]string, error) {
	if len(attributes) == 0 {
		return nil, nil
	}
	normalized := make(map[string]string, len(attributes))
	for key, value := range attributes {
		key = strings.TrimSpace(key)
		if key == "" {
			return nil, fmt.Errorf("attribute key is required")
		}
		if strings.ContainsAny(key, "\r\n") {
			return nil, fmt.Errorf("attribute key %q must not contain line breaks", key)
		}
		if _, exists := normalized[key]; exists {
			return nil, fmt.Errorf("attribute key %q is duplicated after normalization", key)
		}
		normalized[key] = value
	}
	return normalized, nil
}

func normalizeFreshness(metadata Freshness) (Freshness, error) {
	metadata.SourceHash = strings.ToLower(strings.TrimSpace(metadata.SourceHash))
	if err := validateDigest(metadata.SourceHash); err != nil {
		return Freshness{}, fmt.Errorf("freshness source_hash: %w", err)
	}
	var err error
	metadata.ProducedAt, err = normalizeTimestamp(metadata.ProducedAt, "produced_at")
	if err != nil {
		return Freshness{}, err
	}
	if metadata.ValidUntil != "" {
		metadata.ValidUntil, err = normalizeTimestamp(metadata.ValidUntil, "valid_until")
		if err != nil {
			return Freshness{}, err
		}
		produced, _ := time.Parse(time.RFC3339Nano, metadata.ProducedAt)
		validUntil, _ := time.Parse(time.RFC3339Nano, metadata.ValidUntil)
		if !validUntil.After(produced) {
			return Freshness{}, fmt.Errorf("valid_until must be after produced_at")
		}
	}
	return metadata, nil
}

func normalizeTimestamp(value, field string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("freshness %s is required", field)
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return "", fmt.Errorf("freshness %s: %w", field, err)
	}
	return parsed.UTC().Format(time.RFC3339Nano), nil
}

func validateDigest(value string) error {
	if len(value) != sha256.Size*2 {
		return fmt.Errorf("must be a %d-character SHA-256 hex digest", sha256.Size*2)
	}
	if _, err := hex.DecodeString(value); err != nil {
		return fmt.Errorf("must be a SHA-256 hex digest: %w", err)
	}
	return nil
}

func marshalEvidence(evidence Evidence, includeHash bool) ([]byte, error) {
	body := map[string]any{
		"generated_by": evidence.GeneratedBy,
		"freshness": map[string]string{
			"produced_at": evidence.Freshness.ProducedAt,
			"source_hash": evidence.Freshness.SourceHash,
		},
		"id":      evidence.ID,
		"schema":  evidence.Schema,
		"subject": evidence.Subject,
		"type":    evidence.Type,
	}
	if evidence.Freshness.ValidUntil != "" {
		body["freshness"].(map[string]string)["valid_until"] = evidence.Freshness.ValidUntil
	}
	if len(evidence.Attributes) > 0 {
		body["attributes"] = evidence.Attributes
	}
	if includeHash {
		body["hash"] = evidence.Hash
	}
	return json.Marshal(body)
}

func decodeEvidence(raw []byte) (Evidence, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var evidence Evidence
	if err := decoder.Decode(&evidence); err != nil {
		return Evidence{}, &lineError{kind: "invalid-json", err: err}
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return Evidence{}, &lineError{kind: "multiple-values", err: fmt.Errorf("multiple JSON values")}
		}
		return Evidence{}, &lineError{kind: "invalid-json", err: err}
	}
	return evidence, nil
}

func digestBytes(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}
