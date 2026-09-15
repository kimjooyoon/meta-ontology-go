package bidir

import (
	"strings"
	"testing"
)

func TestBXDeltaEvidencePreservesFactOrderAndAtomicState(t *testing.T) {
	first := NewSourcedFact(DeterministicFact, "billing://activity/pay-order", PredicateInvokes, "billing://activity/audit-payment", SourceSpan{File: "a.go", Start: 1, End: 2})
	second := NewSourcedFact(DeterministicFact, "billing://activity/audit-payment", PredicateInvokes, "billing://activity/pay-order", SourceSpan{File: "b.go", Start: 3, End: 4})
	left := FactDelta{Added: FactSet{first, second}}
	right := FactDelta{Added: FactSet{second, first}}
	if factSequenceHash(left) == factSequenceHash(right) || factOrderHash(left) == factOrderHash(right) {
		t.Fatal("delta evidence lost source sequence/order")
	}
	evidence, err := MeasureBXFixture(billingBXFixture{})
	if err != nil {
		t.Fatal(err)
	}
	if !evidence.Delta.ClosureValid || !sameIDs(evidence.Delta.ClosureMembers, evidence.Delta.Locality.Affected) {
		t.Fatalf("delta locality closure was not verified: %#v", evidence.Delta)
	}
	if evidence.Delta.PortOrderHash == "" || evidence.Delta.RelationOrderHash == "" || !strings.Contains(evidence.Delta.CanonicalJSON, "\"candidates\"") {
		t.Fatalf("delta canonical order/candidate schema is incomplete: %#v", evidence.Delta)
	}
	if evidence.RejectedTransaction.Deferred || !evidence.RejectedTransaction.NoWrite || !evidence.PartialConflict.NoWriteObserved || evidence.PartialConflict.RemovedCreated || evidence.PartialConflict.CandidatePromoted {
		t.Fatalf("rejected partial transaction was not observer-proven and non-authoritative: %#v", evidence)
	}
}
func TestBXEvidenceRejectsMissingCanonicalDeltaFields(t *testing.T) {
	evidence, err := MeasureBXFixture(billingBXFixture{})
	if err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(*BXDeltaEvidence){
		"closure-members":    func(delta *BXDeltaEvidence) { delta.ClosureMembers = nil },
		"closure-membership": func(delta *BXDeltaEvidence) { delta.ClosureMembers = []ID{"not-in-closure"} },
		"candidates":         func(delta *BXDeltaEvidence) { delta.Candidates = nil },
		"port-sequence":      func(delta *BXDeltaEvidence) { delta.PortSequence = nil },
		"evidence-hash":      func(delta *BXDeltaEvidence) { delta.EvidenceHash = "" },
	} {
		t.Run(name, func(t *testing.T) {
			copyEvidence := evidence
			mutate(&copyEvidence.Delta)
			if err := copyEvidence.validate(); err == nil {
				t.Fatal("missing canonical delta field was accepted")
			}
		})
	}
}
