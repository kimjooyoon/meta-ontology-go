package analyzer

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
)

// ComparisonCanonical omits host metadata while retaining semantic evidence,
// making equivalent reports comparable across host implementations.
func (r EvidenceReport) ComparisonCanonical() string {
	if !r.Complete() {
		return ""
	}
	records := append([]EvidenceRecord(nil), r.Records...)
	sort.Slice(records, func(i, j int) bool {
		return records[i].comparisonCanonical() < records[j].comparisonCanonical()
	})
	var builder strings.Builder
	builder.WriteString("analyzer-evidence/v1\n")
	for _, record := range records {
		builder.WriteString(record.comparisonCanonical())
		builder.WriteByte('\n')
	}
	return builder.String()
}

// ComparisonDigest returns a stable digest only for a complete report.
func (r EvidenceReport) ComparisonDigest() string {
	canonical := r.ComparisonCanonical()
	if canonical == "" {
		return ""
	}
	digest := sha256.Sum256([]byte(canonical))
	return hex.EncodeToString(digest[:])
}

// GoHostedEvidence projects the current analyzer result into implemented
// reference-host evidence.
func (r Result) GoHostedEvidence() EvidenceReport {
	report := EvidenceReport{Contract: ContractFor(StageGoHosted)}
	for _, fact := range r.Delta.Added {
		report.Records = append(report.Records, EvidenceRecord{
			Kind: EvidenceKindFact, Status: EvidenceStatusDeterministic,
			Subject: fact.Subject, Relation: fact.Relation, Object: fact.Object, Span: fact.Span,
		})
	}
	for _, candidate := range r.Delta.Candidates {
		report.Records = append(report.Records, EvidenceRecord{
			Kind: EvidenceKindCandidate, Status: EvidenceStatusCandidate,
			Subject: candidate.Subject, Relation: candidate.Relation, Reference: candidate.Reference,
			Options: append([]Identity(nil), candidate.Options...), Span: candidate.Span, Reason: candidate.Reason,
		})
	}
	for _, detail := range r.Delta.ImplementationDetails {
		detail = detail.normalized()
		report.Records = append(report.Records, EvidenceRecord{
			Kind: EvidenceKindImplementation, Status: EvidenceStatusImplementation,
			Reference: detail.Reference, Span: detail.Span, Reason: detail.Reason,
			IdentityState: detail.IdentityState,
		})
	}
	sortEvidenceRecords(report.Records)
	return report
}
