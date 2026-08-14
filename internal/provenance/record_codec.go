package provenance

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
)

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
	Type           string            `json:"type"`
	Subject        string            `json:"subject"`
	GeneratedBy    string            `json:"generated_by"`
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
		Schema: e.Schema, ID: e.ID, SemanticID: e.SemanticID,
		Producer: e.Producer, Kind: e.Kind, Status: e.Status,
		SourceSpan: e.SourceSpan, SourceDigest: e.SourceDigest,
		SemanticDigest: e.SemanticDigest, GraphDigest: e.GraphDigest,
		Sequence: e.Sequence, Predecessor: e.Predecessor,
		Attributes: e.Attributes,
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

// CanonicalJSONL returns deterministic LF-terminated records. It does not
// sort records because chain order is semantic and must be preserved.
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
		Schema: wire.Schema, ID: id, SemanticID: wire.SemanticID,
		Producer: wire.Producer, Kind: wire.Kind, Status: wire.Status,
		SourceSpan: SourceSpan{
			URI: wire.SourceSpan.URI, File: wire.SourceSpan.File,
			Start: wire.SourceSpan.Start, End: wire.SourceSpan.End,
		},
		SourceDigest: wire.SourceDigest, SemanticDigest: wire.SemanticDigest,
		GraphDigest: wire.GraphDigest, Sequence: wire.Sequence,
		Predecessor: wire.Predecessor, Attributes: wire.Attributes,
		Freshness: wire.Freshness, Hash: wire.Hash,
		Type: wire.Type, Subject: wire.Subject, GeneratedBy: wire.GeneratedBy,
	}, nil
}

func digestBytes(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}
