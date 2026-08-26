package provenance

import (
	"fmt"
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
