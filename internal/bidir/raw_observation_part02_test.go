package bidir

import (
	"testing"
)

func TestReconcileRawEvidenceDoesNotChangeAuthorityHash(t *testing.T) {
	base, err := Get(billingDocument())
	if err != nil {
		t.Fatal(err)
	}
	first, second := rawEvidenceFact("evidence-a", 30), rawEvidenceFact("evidence-b", 10)
	left, err := Reconcile(base, FactDelta{Added: FactSet{first}})
	if err != nil {
		t.Fatal(err)
	}
	right, err := Reconcile(base, FactDelta{Added: FactSet{second}})
	if err != nil {
		t.Fatal(err)
	}
	if left.RawObservation.EvidenceHash == right.RawObservation.EvidenceHash {
		t.Fatal("distinct evidence records collapsed to one observation hash")
	}
	if SemanticFingerprint(left.Model) != SemanticFingerprint(right.Model) {
		t.Fatal("evidence identity or span changed authoritative semantic hash")
	}
	if !SemanticEquivalent(left.Model, right.Model) {
		t.Fatal("evidence identity or span changed semantic meaning")
	}
}
func rawEvidenceFact(evidenceID string, start int) Fact {
	fact := NewSourcedFact(DeterministicFact, "billing://activity/pay-order", PredicateInvokes, "billing://activity/audit-payment", SourceSpan{File: "raw.gooo", Start: start, End: start + 4})
	fact.EvidenceID = evidenceID
	fact.SubjectKind = ActivityKind
	fact.ObjectKind = ActivityKind
	fact.Attributes = map[string]string{"mode": "observed"}
	return fact
}
func cloneFactDelta(delta FactDelta) FactDelta {
	clone := FactDelta{Added: cloneFacts(delta.Added), Removed: cloneFacts(delta.Removed)}
	return clone
}
func cloneFacts(facts FactSet) FactSet {
	if facts == nil {
		return nil
	}
	clone := make(FactSet, len(facts))
	for index, fact := range facts {
		clone[index] = fact.normalized()
	}
	return clone
}
