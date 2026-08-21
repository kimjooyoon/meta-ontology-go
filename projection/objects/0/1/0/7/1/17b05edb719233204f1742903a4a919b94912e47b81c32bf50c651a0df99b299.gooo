package bidir

import (
	"fmt"
	"strings"
)

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
