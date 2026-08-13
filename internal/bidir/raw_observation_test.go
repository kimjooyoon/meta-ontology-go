package bidir

import (
	"reflect"
	"testing"
)

func TestReconcilePreservesRawDuplicateEvidenceBeforeNormalization(t *testing.T) {
	base, err := Get(billingDocument())
	if err != nil {
		t.Fatal(err)
	}
	first := rawEvidenceFact("evidence-a", 30)
	second := rawEvidenceFact("evidence-b", 10)
	changes := FactDelta{Added: FactSet{first, second}}
	result, err := Reconcile(base, changes)
	if err != nil {
		t.Fatal(err)
	}
	if got := result.RawObservation.Added; len(got) != 2 || got[0].EvidenceID != second.EvidenceID || got[1].EvidenceID != first.EvidenceID {
		t.Fatalf("raw duplicate evidence was not retained deterministically: %#v", got)
	}
	if len(result.Accepted) != 1 || len(result.Model.Relations) != len(base.Relations)+1 {
		t.Fatalf("semantic normalization did not remain deduplicated: result=%#v", result)
	}
	if result.RawObservation.EvidenceHash == "" {
		t.Fatal("raw evidence hash is empty")
	}
}

func TestReconcileRawObservationIsPermutationStableAndDetached(t *testing.T) {
	base, err := Get(billingDocument())
	if err != nil {
		t.Fatal(err)
	}
	first := rawEvidenceFact("evidence-a", 30)
	second := rawEvidenceFact("evidence-b", 10)
	changes := FactDelta{Added: FactSet{first, second}}
	before := cloneFactDelta(changes)
	left, err := Reconcile(base, changes)
	if err != nil {
		t.Fatal(err)
	}
	right, err := Reconcile(base, FactDelta{Added: FactSet{second, first}})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(left.RawObservation, right.RawObservation) {
		t.Fatalf("permuted raw observations differ:\nleft %#v\nright %#v", left.RawObservation, right.RawObservation)
	}
	if left.RawObservation.EvidenceHash != right.RawObservation.EvidenceHash {
		t.Fatal("permuted raw evidence changed evidence hash")
	}
	if SemanticFingerprint(left.Model) != SemanticFingerprint(right.Model) {
		t.Fatal("evidence permutation changed authoritative semantic hash")
	}
	if !reflect.DeepEqual(changes, before) {
		t.Fatalf("reconcile mutated the input observation: got=%#v want=%#v", changes, before)
	}
	left.RawObservation.Added[0].Attributes["mutated"] = "result"
	if changes.Added[0].Attributes["mutated"] != "" {
		t.Fatal("raw observation was not detached from the input")
	}
}

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
