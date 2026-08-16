package bidir

import (
	"fmt"
	"strings"
)

// BXEvidenceRecord keeps one evidence ID paired with its FactKey and span.
// It is observational only and never participates in semantic identity.
type BXEvidenceRecord struct {
	EvidenceID string
	FactKey    string
	Span       SourceSpan
	HasSpan    bool
}

func validEvidenceAuthority(value string) bool {
	return value == "explicit" || value == "derived-non-authoritative" || value == "mixed-non-authoritative"
}

func evidenceAuthority(facts FactSet) string {
	explicit, derived := false, false
	for _, fact := range facts {
		if fact.EvidenceID == "" {
			derived = true
		} else {
			explicit = true
		}
	}
	switch {
	case explicit && derived:
		return "mixed-non-authoritative"
	case derived:
		return "derived-non-authoritative"
	default:
		return "explicit"
	}
}

func evidenceSpanSetHash(evidence BXEvidenceSpanSet) string {
	return digest(evidenceSpanSetCanonical(evidence))
}

func evidenceSpanSetCanonical(evidence BXEvidenceSpanSet) string {
	var builder strings.Builder
	writePart(&builder, evidence.EvidenceIDAuthority)
	fmt.Fprintf(&builder, "%d|%d|", evidence.IDCount, evidence.SpanCount)
	for _, id := range evidence.IDs {
		writePart(&builder, id)
	}
	for _, key := range evidence.FactKeys {
		writePart(&builder, key)
	}
	for _, span := range evidence.Spans {
		writeSpan(&builder, span)
	}
	for _, record := range evidence.Records {
		writePart(&builder, record.EvidenceID)
		writePart(&builder, record.FactKey)
		fmt.Fprintf(&builder, "%t|", record.HasSpan)
		writeSpan(&builder, record.Span)
	}
	return builder.String()
}

func validateEvidenceRecords(evidence BXEvidenceSpanSet) error {
	if len(evidence.Records) != evidence.IDCount || len(evidence.Records) != len(evidence.FactKeys) {
		return fmt.Errorf("evidence record cardinality does not match IDs and FactKeys")
	}
	spanCount := 0
	spanIndex := 0
	for index, record := range evidence.Records {
		if record.EvidenceID != evidence.IDs[index] || record.FactKey != evidence.FactKeys[index] {
			return fmt.Errorf("evidence record %d is not paired with its ID and FactKey", index)
		}
		if record.HasSpan {
			if !record.Span.Valid() {
				return fmt.Errorf("evidence record %d marks an invalid span", index)
			}
			if spanIndex >= len(evidence.Spans) || evidence.Spans[spanIndex] != record.Span {
				return fmt.Errorf("evidence record %d is not paired with its span", index)
			}
			spanIndex++
			spanCount++
		} else if record.Span.Valid() {
			return fmt.Errorf("evidence record %d has an unmarked span", index)
		}
	}
	if spanCount != evidence.SpanCount {
		return fmt.Errorf("evidence record span cardinality does not match spans")
	}
	if spanIndex != len(evidence.Spans) {
		return fmt.Errorf("evidence span list has unpaired spans")
	}
	return nil
}
