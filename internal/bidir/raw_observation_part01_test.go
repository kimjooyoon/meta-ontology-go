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
