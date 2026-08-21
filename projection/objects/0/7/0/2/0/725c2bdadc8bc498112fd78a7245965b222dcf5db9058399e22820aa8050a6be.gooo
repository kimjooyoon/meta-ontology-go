package analyzer

import (
	"sort"
	"strconv"
	"strings"
)

// GoooHostedEvidence is an explicit deferred report until a gooo-hosted
// analyzer and independent comparison gate exist.
func (r Result) GoooHostedEvidence() EvidenceReport {
	return EvidenceReport{
		Contract: ContractFor(StageGoooHosted),
		Reason:   "gooo-hosted analyzer is not implemented",
	}
}
func sortEvidenceRecords(records []EvidenceRecord) {
	sort.SliceStable(records, func(i, j int) bool {
		return records[i].comparisonCanonical() < records[j].comparisonCanonical()
	})
}
func (e EvidenceRecord) comparisonCanonical() string {
	var builder strings.Builder
	writeEvidenceField(&builder, string(e.Kind))
	writeEvidenceField(&builder, string(e.Status))
	writeEvidenceField(&builder, e.Subject.Namespace)
	writeEvidenceField(&builder, e.Subject.ID)
	writeEvidenceField(&builder, string(e.Relation))
	writeEvidenceField(&builder, e.Object.Namespace)
	writeEvidenceField(&builder, e.Object.ID)
	writeEvidenceField(&builder, e.Reference)
	options := append([]Identity(nil), e.Options...)
	sort.Slice(options, func(i, j int) bool { return identityLess(options[i], options[j]) })
	for _, option := range options {
		writeEvidenceField(&builder, option.Namespace)
		writeEvidenceField(&builder, option.ID)
	}
	writeEvidenceField(&builder, e.Span.Filename)
	writeEvidenceField(&builder, strconv.Itoa(e.Span.Start.Offset))
	writeEvidenceField(&builder, strconv.Itoa(e.Span.Start.Line))
	writeEvidenceField(&builder, strconv.Itoa(e.Span.Start.Column))
	writeEvidenceField(&builder, strconv.Itoa(e.Span.End.Offset))
	writeEvidenceField(&builder, strconv.Itoa(e.Span.End.Line))
	writeEvidenceField(&builder, strconv.Itoa(e.Span.End.Column))
	writeEvidenceField(&builder, e.Reason)
	writeEvidenceField(&builder, string(e.IdentityState))
	return builder.String()
}
func validIdentityOptions(options []Identity) bool {
	if len(options) == 0 {
		return false
	}
	for _, option := range options {
		if !option.Valid() {
			return false
		}
	}
	return true
}
func evidenceSpanValid(span Span) bool {
	return span.Filename != "" && span.Start.Offset >= 0 && span.End.Offset > span.Start.Offset && span.Start.Line > 0 && span.End.Line > 0 && span.Start.Column > 0 && span.End.Column > 0
}
func writeEvidenceField(builder *strings.Builder, value string) {
	builder.WriteString(strconv.Itoa(len(value)))
	builder.WriteByte(':')
	builder.WriteString(value)
	builder.WriteByte('|')
}
