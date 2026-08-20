package bidir

import (
	"fmt"
)

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
