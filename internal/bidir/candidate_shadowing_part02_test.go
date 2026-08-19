package bidir

import (
	"reflect"
	"testing"
)

func TestCandidateShadowingIsPermutationStableAndRejectsTamperedModel(t *testing.T) {
	base, err := Get(billingDocument())
	if err != nil {
		t.Fatal(err)
	}
	candidate := NewSourcedFact(CandidateFact, "billing://entity/payment", PredicateWasDerivedFrom, "billing://entity/order", SourceSpan{File: "candidate.go", Start: 3, End: 4})
	deterministic := NewSourcedFact(DeterministicFact, candidate.Subject, candidate.Predicate, candidate.Object, SourceSpan{File: "accepted.go", Start: 5, End: 6})
	leftChanges := FactDelta{Added: FactSet{candidate, deterministic}}
	rightChanges := FactDelta{Added: FactSet{deterministic, candidate}}
	leftBefore, rightBefore := cloneFactDelta(leftChanges), cloneFactDelta(rightChanges)
	left, err := Reconcile(base, leftChanges)
	if err != nil {
		t.Fatal(err)
	}
	right, err := Reconcile(base, rightChanges)
	if err != nil {
		t.Fatal(err)
	}
	if !SemanticEquivalent(left.Model, right.Model) || len(left.Model.Candidates) != 0 || len(right.Model.Candidates) != 0 {
		t.Fatalf("candidate shadowing was not permutation-stable: left=%#v right=%#v", left.Model, right.Model)
	}
	if left.RawObservation.EvidenceHash != right.RawObservation.EvidenceHash {
		t.Fatal("candidate permutation changed raw evidence hash")
	}
	if !reflect.DeepEqual(leftChanges, leftBefore) || !reflect.DeepEqual(rightChanges, rightBefore) {
		t.Fatal("candidate shadowing mutated an input delta")
	}
	tampered := left.Model.Clone()
	tampered.Candidates = FactSet{candidate}
	if err := tampered.Validate(); err == nil {
		t.Fatal("tampered model retained an overlapping candidate")
	}
}
