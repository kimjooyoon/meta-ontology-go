package bidir

import (
	"sort"
)

func evidenceSpans(facts FactSet) BXEvidenceSpanSet {
	observed := make(FactSet, len(facts))
	for index, fact := range facts {
		observed[index] = fact.normalized()
	}
	sort.SliceStable(observed, func(i, j int) bool { return rawFactLess(observed[i], observed[j]) })
	ids := make([]string, len(observed))
	factKeys := make([]string, len(observed))
	records := make([]BXEvidenceRecord, len(observed))
	spans := make([]SourceSpan, 0, len(observed))
	occurrences := make(map[string]int, len(observed))
	for index, fact := range observed {
		canonical := factCanonical(fact)
		ids[index] = factEvidenceID(fact, occurrences[canonical])
		occurrences[canonical]++
		factKeys[index] = factID(fact)
		records[index] = BXEvidenceRecord{EvidenceID: ids[index], FactKey: factKeys[index], Span: fact.Source, HasSpan: fact.Source.Valid()}
		if fact.Source.Valid() {
			spans = append(spans, fact.Source)
		}
	}
	evidence := BXEvidenceSpanSet{IDs: ids, FactKeys: factKeys, Spans: spans, Records: records, IDCount: len(ids), SpanCount: len(spans), EvidenceIDAuthority: evidenceAuthority(observed)}
	evidence.Hash = evidenceSpanSetHash(evidence)
	return evidence
}
func spanLess(left, right SourceSpan) bool {
	if left.File != right.File {
		return left.File < right.File
	}
	if left.Start != right.Start {
		return left.Start < right.Start
	}
	if left.End != right.End {
		return left.End < right.End
	}
	if left.StartLine != right.StartLine {
		return left.StartLine < right.StartLine
	}
	if left.StartColumn != right.StartColumn {
		return left.StartColumn < right.StartColumn
	}
	if left.EndLine != right.EndLine {
		return left.EndLine < right.EndLine
	}
	return left.EndColumn < right.EndColumn
}
func artifact(hash string, count int) BXArtifactEvidence {
	return BXArtifactEvidence{Hash: hash, Count: count}
}
