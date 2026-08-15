package provenance

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

func normalizeEvidence(evidence Evidence) (Evidence, error) {
	if evidence.Schema == 0 {
		evidence.Schema = SchemaVersion
	}
	if evidence.Schema != SchemaVersion {
		return Evidence{}, fmt.Errorf("unsupported schema %d", evidence.Schema)
	}
	if err := normalizeIdentity(&evidence); err != nil {
		return Evidence{}, err
	}
	if err := normalizeClassification(&evidence); err != nil {
		return Evidence{}, err
	}
	if err := normalizeDigests(&evidence); err != nil {
		return Evidence{}, err
	}
	if err := normalizeEvidenceMetadata(&evidence); err != nil {
		return Evidence{}, err
	}
	if err := normalizePredecessor(&evidence); err != nil {
		return Evidence{}, err
	}
	return evidence, nil
}

func normalizeIdentity(evidence *Evidence) error {
	var err error
	evidence.ID, err = normalizeIdentifier(evidence.ID, "id")
	if err != nil {
		return err
	}
	if evidence.Producer == "" {
		evidence.Producer = evidence.GeneratedBy
	}
	evidence.Producer, err = normalizeIdentifier(evidence.Producer, "producer")
	if err != nil {
		return err
	}
	if evidence.SemanticID == "" {
		evidence.SemanticID = evidence.Subject
	}
	if evidence.SemanticID == "" {
		evidence.SemanticID = evidence.ID
	}
	evidence.SemanticID, err = normalizeIdentifier(evidence.SemanticID, "semantic_id")
	return err
}

func normalizeClassification(evidence *Evidence) error {
	if evidence.Kind == "" {
		evidence.Kind = EvidenceKind(evidence.Type)
	}
	kind, err := normalizeIdentifier(string(evidence.Kind), "kind")
	if err != nil {
		return err
	}
	evidence.Kind = EvidenceKind(kind)
	if evidence.Status == "" {
		if value, ok := evidence.Attributes["status"]; ok {
			evidence.Status = EvidenceStatus(value)
		}
	}
	if evidence.Status == "" {
		return fmt.Errorf("status is required")
	}
	attributeStatus, hasAttributeStatus := evidence.Attributes["status"]
	evidence.Status = normalizeStatus(evidence.Status)
	if hasAttributeStatus && normalizeStatus(EvidenceStatus(attributeStatus)) != evidence.Status {
		return fmt.Errorf("attributes.status does not match status")
	}
	if !validStatus(evidence.Status) {
		return fmt.Errorf("unsupported status %q", evidence.Status)
	}
	return nil
}

func normalizeDigests(evidence *Evidence) error {
	var err error
	if evidence.SourceDigest == "" {
		evidence.SourceDigest = evidence.Freshness.SourceHash
	}
	evidence.SourceDigest, err = normalizeDigest(evidence.SourceDigest, "source_digest")
	if err != nil {
		return err
	}
	if evidence.Freshness.SourceHash != "" {
		freshness, freshnessErr := normalizeDigest(evidence.Freshness.SourceHash, "freshness.source_hash")
		if freshnessErr != nil {
			return freshnessErr
		}
		if freshness != evidence.SourceDigest {
			return fmt.Errorf("freshness source_hash does not match source_digest")
		}
	}
	evidence.Freshness.SourceHash = evidence.SourceDigest
	evidence.SemanticDigest, err = normalizeDigest(evidence.SemanticDigest, "semantic_digest")
	if err != nil {
		return err
	}
	evidence.GraphDigest, err = normalizeDigest(evidence.GraphDigest, "graph_digest")
	return err
}

func normalizeEvidenceMetadata(evidence *Evidence) error {
	var err error
	evidence.SourceSpan, err = normalizeSourceSpan(evidence.SourceSpan)
	if err != nil {
		return err
	}
	evidence.Attributes, err = normalizeAttributes(evidence.Attributes)
	if err != nil {
		return err
	}
	evidence.Freshness, err = normalizeFreshness(evidence.Freshness)
	return err
}

func normalizePredecessor(evidence *Evidence) error {
	if evidence.Predecessor == nil {
		return nil
	}
	link := *evidence.Predecessor
	var err error
	link.ID, err = normalizeIdentifier(link.ID, "predecessor.id")
	if err != nil {
		return err
	}
	link.Digest, err = normalizeDigest(link.Digest, "predecessor.digest")
	if err != nil {
		return err
	}
	evidence.Predecessor = &link
	return nil
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
