package analyzer

// AuthoritativeFacts returns only deterministic semantic facts. The returned
// records are copies so candidate and implementation views remain separate.
func (r EvidenceReport) AuthoritativeFacts() []EvidenceRecord {
	return filterEvidence(r.Records, EvidenceKindFact)
}

// CandidateEvidence returns unresolved semantic relations with all options.
func (r EvidenceReport) CandidateEvidence() []EvidenceRecord {
	return filterEvidence(r.Records, EvidenceKindCandidate)
}

// ImplementationEvidence returns Go details that have not crossed the
// semantic boundary.
func (r EvidenceReport) ImplementationEvidence() []EvidenceRecord {
	return filterEvidence(r.Records, EvidenceKindImplementation)
}

func filterEvidence(records []EvidenceRecord, kind EvidenceKind) []EvidenceRecord {
	filtered := make([]EvidenceRecord, 0)
	for _, record := range records {
		if record.Kind == kind {
			filtered = append(filtered, copyEvidenceRecord(record))
		}
	}
	return filtered
}

func copyEvidenceRecord(record EvidenceRecord) EvidenceRecord {
	record.Options = append([]Identity(nil), record.Options...)
	return record
}
