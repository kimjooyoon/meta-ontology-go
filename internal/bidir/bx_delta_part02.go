package bidir

type canonicalDeltaEvidence struct {
	SequenceHash        string                    `json:"sequence_hash"`
	OrderHash           string                    `json:"order_hash"`
	Added               []string                  `json:"added"`
	Removed             []string                  `json:"removed"`
	Candidates          []string                  `json:"candidates"`
	PortSequence        []string                  `json:"port_sequence"`
	RelationSequence    []string                  `json:"relation_sequence"`
	Touched             []ID                      `json:"touched"`
	Affected            []ID                      `json:"affected"`
	ClosureMembers      []ID                      `json:"closure_members"`
	ClosureHash         string                    `json:"closure_hash"`
	EvidenceIDs         []string                  `json:"evidence_ids"`
	EvidenceFactKeys    []string                  `json:"evidence_fact_keys"`
	EvidenceSpans       []string                  `json:"evidence_spans"`
	EvidenceRecords     []canonicalEvidenceRecord `json:"evidence_records"`
	EvidenceIDCount     int                       `json:"evidence_id_count"`
	EvidenceSpanCount   int                       `json:"evidence_span_count"`
	EvidenceIDAuthority string                    `json:"evidence_id_authority"`
	EvidenceHash        string                    `json:"evidence_hash"`
	Partial             bool                      `json:"partial_observation"`
}
type canonicalEvidenceRecord struct {
	EvidenceID string `json:"evidence_id"`
	FactKey    string `json:"fact_key"`
	Span       string `json:"span"`
	HasSpan    bool   `json:"has_span"`
}

func deltaJSON(delta FactDelta, evidence BXDeltaEvidence) string {
	value := canonicalDeltaEvidence{
		SequenceHash:        evidence.SequenceHash,
		OrderHash:           evidence.OrderHash,
		Added:               factCanonicalValues(delta.Added),
		Removed:             factCanonicalValues(delta.Removed),
		Candidates:          evidence.Candidates,
		PortSequence:        evidence.PortSequence,
		RelationSequence:    evidence.RelationSequence,
		Touched:             evidence.Locality.Touched,
		Affected:            evidence.Locality.Affected,
		ClosureMembers:      evidence.ClosureMembers,
		ClosureHash:         evidence.LocalityClosureHash,
		EvidenceIDs:         evidence.EvidenceSpans.IDs,
		EvidenceFactKeys:    evidence.EvidenceSpans.FactKeys,
		EvidenceSpans:       spanTexts(evidence.EvidenceSpans.Spans),
		EvidenceRecords:     canonicalEvidenceRecords(evidence.EvidenceSpans.Records),
		EvidenceIDCount:     evidence.EvidenceSpans.IDCount,
		EvidenceSpanCount:   evidence.EvidenceSpans.SpanCount,
		EvidenceIDAuthority: evidence.EvidenceSpans.EvidenceIDAuthority,
		EvidenceHash:        evidence.EvidenceHash,
		Partial:             evidence.PartialObservation,
	}
	result, _ := canonicalJSON(value)
	return result
}
func canonicalEvidenceRecords(records []BXEvidenceRecord) []canonicalEvidenceRecord {
	values := make([]canonicalEvidenceRecord, len(records))
	for index, record := range records {
		values[index] = canonicalEvidenceRecord{EvidenceID: record.EvidenceID, FactKey: record.FactKey, Span: spanText(record.Span), HasSpan: record.HasSpan}
	}
	return values
}
func localityJSON(locality Locality, closureHash string) string {
	value := struct {
		Touched  []ID   `json:"touched"`
		Affected []ID   `json:"affected"`
		Members  []ID   `json:"closure_members"`
		Closure  string `json:"closure_hash"`
	}{Touched: append([]ID{}, locality.Touched...), Affected: append([]ID{}, locality.Affected...), Members: append([]ID{}, locality.Affected...), Closure: closureHash}
	result, _ := canonicalJSON(value)
	return result
}
