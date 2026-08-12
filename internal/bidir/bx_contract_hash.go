package bidir

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

func digest(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func documentDigest(document Document) string {
	return digest(documentCanonical(document))
}

func documentCanonical(document Document) string {
	var builder strings.Builder
	writePart(&builder, document.Package)
	writePart(&builder, document.Namespace)
	for _, declaration := range document.Declarations {
		writePart(&builder, string(declaration.Kind))
		writePart(&builder, string(declaration.ID))
		writePart(&builder, declaration.Name)
		writeMapFingerprint(&builder, declaration.Attributes)
		writeSpan(&builder, declaration.Span)
		writeReferences(&builder, declaration.Inputs)
		writeReferences(&builder, declaration.Outputs)
	}
	for _, relation := range document.Relations {
		writePart(&builder, string(relation.Kind))
		writePart(&builder, string(relation.Source))
		writePart(&builder, string(relation.Target))
		writeMapFingerprint(&builder, relation.Attributes)
		writeSpan(&builder, relation.Span)
	}
	return builder.String()
}

func writeReferences(builder *strings.Builder, references []Reference) {
	for _, reference := range references {
		writePart(builder, string(reference.ID))
		writePart(builder, reference.Name)
		writePart(builder, reference.Namespace)
		writeSpan(builder, reference.Span)
	}
}

func writeSpan(builder *strings.Builder, span SourceSpan) {
	writePart(builder, span.File)
	fmt.Fprintf(builder, "%d,%d,%d,%d,%d,%d;", span.Start, span.End, span.StartLine, span.StartColumn, span.EndLine, span.EndColumn)
}

func writePart(builder *strings.Builder, value string) {
	fmt.Fprintf(builder, "%d:", len(value))
	builder.WriteString(value)
}

func factCanonical(fact Fact) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "%d|", fact.Layer)
	writePart(&builder, string(fact.Subject))
	writePart(&builder, string(fact.Predicate))
	writePart(&builder, string(fact.Object))
	writePart(&builder, string(fact.SubjectKind))
	writePart(&builder, string(fact.ObjectKind))
	writeMapFingerprint(&builder, fact.Attributes)
	writeSpan(&builder, fact.Source)
	writePart(&builder, fact.Reason)
	return builder.String()
}

func factID(fact Fact) string {
	return fmt.Sprintf("%d:%s:%s:%s", fact.Layer, fact.Subject, fact.Predicate, fact.Object)
}

func factEvidenceID(fact Fact, occurrence int) string {
	if fact.EvidenceID != "" {
		return fact.EvidenceID
	}
	return "urn:gooo:evidence:" + digest(fmt.Sprintf("%s|%d", factCanonical(fact), occurrence))
}

func factSequenceHash(delta FactDelta) string {
	var builder strings.Builder
	writeFacts(&builder, "added", delta.Added)
	writeFacts(&builder, "removed", delta.Removed)
	return digest(builder.String())
}

func factOrderHash(delta FactDelta) string {
	var builder strings.Builder
	writeFactIDs(&builder, "added", delta.Added)
	writeFactIDs(&builder, "removed", delta.Removed)
	return digest(builder.String())
}

func writeFacts(builder *strings.Builder, label string, facts FactSet) {
	writePart(builder, label)
	for index, fact := range facts {
		fmt.Fprintf(builder, "%d|", index)
		writePart(builder, factCanonical(fact))
	}
}

func writeFactIDs(builder *strings.Builder, label string, facts FactSet) {
	writePart(builder, label)
	for index, fact := range facts {
		fmt.Fprintf(builder, "%d|", index)
		writePart(builder, factID(fact))
	}
}

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

func baseEvidence(input BXBaseEvidenceInput, document Document, model Model) (BXBaseEvidence, error) {
	if documentDigest(input.DSL) != documentDigest(document) {
		return BXBaseEvidence{}, fmt.Errorf("base DSL artifact does not match fixture document")
	}
	if !SemanticEquivalent(input.IR, model) {
		return BXBaseEvidence{}, fmt.Errorf("base IR artifact does not match fixture model")
	}
	if len(input.Go) == 0 || len(input.SourceMap) == 0 || len(input.Evidence) == 0 || len(input.Provenance) == 0 {
		return BXBaseEvidence{}, fmt.Errorf("base evidence requires non-empty Go, source-map, evidence, and provenance artifacts")
	}
	return BXBaseEvidence{
		DSL:        artifact(documentDigest(input.DSL), 1),
		IR:         artifact(SemanticFingerprint(input.IR), len(input.IR.Nodes)+len(input.IR.Relations)),
		Go:         artifact(digestFacts(input.Go), len(input.Go)),
		SourceMap:  artifact(digestSpans(input.SourceMap), len(input.SourceMap)),
		Evidence:   artifact(digestFacts(input.Evidence), len(input.Evidence)),
		Provenance: artifact(digestSpans(input.Provenance), len(input.Provenance)),
	}, nil
}

func digestFacts(facts FactSet) string {
	copySet := facts.Normalized()
	var builder strings.Builder
	for _, fact := range copySet {
		writePart(&builder, factCanonical(fact))
	}
	return digest(builder.String())
}

func digestSpans(spans []SourceSpan) string {
	copySpans := append([]SourceSpan(nil), spans...)
	sort.Slice(copySpans, func(i, j int) bool { return spanLess(copySpans[i], copySpans[j]) })
	var builder strings.Builder
	for _, span := range copySpans {
		writeSpan(&builder, span)
	}
	return digest(builder.String())
}
