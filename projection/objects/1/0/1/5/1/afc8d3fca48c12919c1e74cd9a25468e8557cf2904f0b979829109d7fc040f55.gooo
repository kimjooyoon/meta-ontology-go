package bidir

import (
	"testing"
)

func TestBXDeltaEvidenceRejectsTamperedEvidenceRecord(t *testing.T) {
	evidence, err := MeasureBXFixture(billingBXFixture{})
	if err != nil {
		t.Fatal(err)
	}
	copyEvidence := evidence
	copyEvidence.Delta.EvidenceSpans.Records = append([]BXEvidenceRecord(nil), evidence.Delta.EvidenceSpans.Records...)
	copyEvidence.Delta.EvidenceSpans.Records[0].FactKey = "tampered"
	if err := copyEvidence.validate(); err == nil {
		t.Fatal("tampered evidence record was accepted")
	}
}
func duplicateEvidenceFact(evidenceID string, start int) Fact {
	fact := NewSourcedFact(DeterministicFact, "billing://activity/pay-order", PredicateInvokes, "billing://activity/audit-payment", SourceSpan{File: "duplicate.gooo", Start: start, End: start + 4})
	fact.EvidenceID = evidenceID
	fact.SubjectKind = ActivityKind
	fact.ObjectKind = ActivityKind
	return fact
}
