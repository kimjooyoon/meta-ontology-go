package query

import "encoding/json"

// CanonicalJSON omits the self-referential hash and emits stable response
// fields. Callers verify Hash by hashing these bytes.
func (response Response) CanonicalJSON() ([]byte, error) {
	canonical := response
	canonical.Hash = ""
	return json.Marshal(canonical)
}

// CanonicalDigest returns the response view receipt digest.
func (response Response) CanonicalDigest() (string, error) {
	canonical, err := response.CanonicalJSON()
	if err != nil {
		return "", err
	}
	return digestBytes(canonical), nil
}

func (response *Response) seal() error {
	digest, err := response.CanonicalDigest()
	if err != nil {
		return err
	}
	response.Hash = digest
	return nil
}

func envelopeMetadata(metadata ProjectionMetadata) EnvelopeMetadata {
	irStatus := "unavailable"
	if metadata.SemanticDigest != "" {
		irStatus = "available"
	}
	provenanceStatus := metadata.ProvenanceStatus
	if provenanceStatus == "" || provenanceStatus == "unknown" || provenanceStatus == "known_empty" {
		provenanceStatus = StatusDeferred
	}
	labels := append([]AuthorityLabel(nil), metadata.AuthorityLabels...)
	for index := range labels {
		if labels[index].View == "provenance" && provenanceStatus == StatusDeferred {
			labels[index].Status = StatusDeferred
		}
	}
	return EnvelopeMetadata{
		SchemaVersion:     metadata.SchemaVersion,
		Namespace:         metadata.Namespace,
		GraphHash:         metadata.GraphHash,
		SemanticDigest:    metadata.SemanticDigest,
		ProjectionStatus:  metadata.ProjectionStatus,
		SourceStatus:      metadata.SourceStatus,
		IRStatus:          irStatus,
		EvidenceStatus:    metadata.EvidenceStatus,
		ProvenanceStatus:  provenanceStatus,
		DerivedStatus:     metadata.DerivedStatus,
		DerivedRuleSchema: metadata.DerivedRuleSchema,
		DerivedRuleDigest: metadata.DerivedRuleDigest,
		AuthorityLabels:   labels,
	}
}
