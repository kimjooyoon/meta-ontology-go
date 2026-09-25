package provenance

import (
	"fmt"
)

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
