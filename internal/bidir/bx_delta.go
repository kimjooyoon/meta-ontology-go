package bidir

import "fmt"

func makeDeltaEvidence(delta FactDelta, locality Locality, partial bool, base, after Model) (BXDeltaEvidence, error) {
	evidence := makeDeltaEvidenceUnchecked(delta, locality, partial, base, after)
	if evidence.CanonicalJSON == "" || evidence.LocalityCanonicalJSON == "" {
		return BXDeltaEvidence{}, fmt.Errorf("canonical delta or locality JSON is empty")
	}
	return evidence, nil
}

func makeDeltaEvidenceUnchecked(delta FactDelta, locality Locality, partial bool, base, after Model) BXDeltaEvidence {
	// Capture raw observations before Reconcile normalizes FactKey duplicates;
	// semantic application and evidence retention are separate boundaries.
	facts := append(append(FactSet{}, delta.Added...), delta.Removed...)
	ports, relations := orderedSequences(after)
	portHash, relationHash := sequenceHash(ports), sequenceHash(relations)
	sequenceHash := factSequenceHash(delta)
	closure := LocalityBetween(base, after)
	evidenceSet := evidenceSpans(facts)
	evidence := BXDeltaEvidence{
		SequenceHash:        sequenceHash,
		OrderHash:           deltaOrderHash(sequenceHash, portHash, relationHash),
		Locality:            detachedLocality(locality),
		Added:               factCanonicalValues(delta.Added),
		Removed:             factCanonicalValues(delta.Removed),
		LocalityClosureHash: localityDigest(locality),
		ClosureMembers:      append([]ID{}, locality.Affected...),
		ClosureValid:        sameLocality(locality, closure),
		Candidates:          factCanonicalValues(candidateFacts(facts)),
		PortSequence:        ports,
		RelationSequence:    relations,
		PortOrderHash:       portHash,
		RelationOrderHash:   relationHash,
		EvidenceSpans:       evidenceSet,
		EvidenceHash:        evidenceSet.Hash,
		PartialObservation:  partial,
		RemovedCreated:      removedCreated(base, after, delta),
		CandidatePromoted:   candidatePromoted(base, delta, after),
	}
	evidence.CanonicalJSON = deltaJSON(delta, evidence)
	evidence.LocalityCanonicalJSON = localityJSON(locality, evidence.LocalityClosureHash)
	return evidence
}

// deltaOrderHash binds the observed fact sequence to the source-authoritative
// semantic collection orders. All inputs are fixed-width SHA-256 values, so
// the delimiter is unambiguous and the hash can be recomputed from the
// detached evidence record alone.
func deltaOrderHash(sequenceHash, portHash, relationHash string) string {
	return digest(sequenceHash + "|" + portHash + "|" + relationHash)
}

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

func factCanonicalValues(facts FactSet) []string {
	values := make([]string, len(facts))
	for index, fact := range facts {
		values[index] = factCanonical(fact)
	}
	return values
}

func candidateFacts(facts FactSet) FactSet {
	candidates := make(FactSet, 0)
	for _, fact := range facts {
		if fact.Layer == CandidateFact {
			candidates = append(candidates, fact)
		}
	}
	return candidates
}

func spanTexts(spans []SourceSpan) []string {
	texts := make([]string, len(spans))
	for index, span := range spans {
		texts[index] = spanText(span)
	}
	return texts
}

func sameLocality(left, right Locality) bool {
	return sameIDs(left.Touched, right.Touched) && sameIDs(left.Affected, right.Affected)
}

func detachedLocality(locality Locality) Locality {
	return Locality{Touched: append([]ID(nil), locality.Touched...), Affected: append([]ID(nil), locality.Affected...)}
}
