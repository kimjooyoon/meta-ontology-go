package provenance

import (
	"bytes"
	"encoding/json"
	"fmt"
)

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
