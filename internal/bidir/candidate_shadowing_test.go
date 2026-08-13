package bidir

import (
	"reflect"
	"testing"
)

func TestDeterministicFactShadowsCandidateWithoutLosingObservation(t *testing.T) {
	base, err := Get(billingDocument())
	if err != nil {
		t.Fatal(err)
	}
	candidate := NewSourcedFact(CandidateFact, "billing://entity/payment", PredicateWasDerivedFrom, "billing://entity/order", SourceSpan{File: "candidate.go", Start: 3, End: 4})
	deterministic := NewSourcedFact(DeterministicFact, "billing://entity/payment", PredicateWasDerivedFrom, "billing://entity/order", SourceSpan{File: "accepted.go", Start: 5, End: 6})
	result, err := Reconcile(base, FactDelta{Added: FactSet{candidate, deterministic}})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.RawObservation.Added) != 2 || !result.RawObservation.Added.Contains(candidate) {
		t.Fatalf("candidate observation was not retained in raw evidence: %#v", result.RawObservation.Added)
	}
	if result.Model.Candidates.Contains(candidate) {
		t.Fatal("candidate remained in authoritative model")
	}
	if _, exists := findRelation(result.Model, deterministic.Predicate, deterministic.Subject, deterministic.Object); !exists {
		t.Fatal("deterministic fact was not accepted")
	}
}

func TestCandidateAddedAfterDeterministicFactIsShadowedAndDetached(t *testing.T) {
	base, err := Get(billingDocument())
	if err != nil {
		t.Fatal(err)
	}
	deterministic := NewSourcedFact(DeterministicFact, "billing://entity/payment", PredicateWasDerivedFrom, "billing://entity/order", SourceSpan{File: "accepted.go", Start: 5, End: 6})
	accepted, err := Reconcile(base, FactDelta{Added: FactSet{deterministic}})
	if err != nil {
		t.Fatal(err)
	}
	candidate := NewSourcedFact(CandidateFact, deterministic.Subject, deterministic.Predicate, deterministic.Object, SourceSpan{File: "candidate.go", Start: 3, End: 4})
	candidate.Attributes = map[string]string{"observed": "true"}
	before := candidate.normalized()
	result, err := Reconcile(accepted.Model, FactDelta{Added: FactSet{candidate}})
	if err != nil {
		t.Fatal(err)
	}
	if !result.RawObservation.Added.Contains(candidate) || result.Model.Candidates.Contains(candidate) {
		t.Fatalf("shadowed candidate boundary is inconsistent: result=%#v", result)
	}
	if !SemanticEquivalent(accepted.Model, result.Model) || !reflect.DeepEqual(candidate, before) {
		t.Fatal("shadowed candidate changed semantic state or mutated input")
	}
}

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
