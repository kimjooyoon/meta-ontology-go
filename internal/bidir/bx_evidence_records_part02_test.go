package bidir

import (
	"reflect"
	"testing"
)

func TestBXDeltaEvidencePreservesOrderAndCanonicalizesModelPermutation(t *testing.T) {
	base, err := Get(billingDocument())
	if err != nil {
		t.Fatal(err)
	}
	relation := Relation{Kind: PredicateInvokes, Source: "billing://activity/pay-order", Target: "billing://activity/audit-payment", Span: SourceSpan{File: "order.gooo", Start: 4, End: 8}}
	after, err := base.Apply(Delta{AddedRelations: []Relation{relation}})
	if err != nil {
		t.Fatal(err)
	}
	permuted := after.Clone()
	reverseModelCollections(&permuted)
	locality := LocalityBetween(base, after)
	changes := FactDelta{Added: FactSet{duplicateEvidenceFact("evidence-a", 10), duplicateEvidenceFact("evidence-b", 20)}}
	left := makeDeltaEvidenceUnchecked(changes, locality, false, base, after)
	right := makeDeltaEvidenceUnchecked(FactDelta{Added: FactSet{changes.Added[1], changes.Added[0]}}, locality, false, base, permuted)
	if left.SequenceHash == right.SequenceHash {
		t.Fatal("fact source sequence hash ignored observation order")
	}
	if left.OrderHash == right.OrderHash {
		t.Fatal("canonical order hash ignored observation order")
	}
	if err := validateDeltaEvidence(left); err != nil {
		t.Fatalf("left permutation evidence failed self-consistency: %v", err)
	}
	if err := validateDeltaEvidence(right); err != nil {
		t.Fatalf("right permutation evidence failed self-consistency: %v", err)
	}
	if left.EvidenceHash != right.EvidenceHash || !reflect.DeepEqual(left.EvidenceSpans, right.EvidenceSpans) {
		t.Fatal("permuted duplicate evidence changed canonical evidence boundary")
	}
	if left.PortOrderHash != right.PortOrderHash || left.RelationOrderHash != right.RelationOrderHash || !reflect.DeepEqual(left.RelationSequence, right.RelationSequence) {
		t.Fatal("model collection permutation changed canonical source order")
	}
}
func TestBXDeltaEvidenceDoesNotMutateInputsOrShareLocality(t *testing.T) {
	base, err := Get(billingDocument())
	if err != nil {
		t.Fatal(err)
	}
	relation := Relation{Kind: PredicateInvokes, Source: "billing://activity/pay-order", Target: "billing://activity/audit-payment", Span: SourceSpan{File: "no-write.gooo", Start: 4, End: 8}}
	after, err := base.Apply(Delta{AddedRelations: []Relation{relation}})
	if err != nil {
		t.Fatal(err)
	}
	locality := LocalityBetween(base, after)
	changes := FactDelta{Added: FactSet{duplicateEvidenceFact("evidence-a", 10)}}
	beforeChanges := cloneFactDelta(changes)
	beforeLocality := detachedLocality(locality)
	evidence := makeDeltaEvidenceUnchecked(changes, locality, false, base, after)
	evidence.Locality.Touched[0] = "mutated"
	evidence.EvidenceSpans.Records[0].EvidenceID = "mutated"
	if !reflect.DeepEqual(changes, beforeChanges) || !reflect.DeepEqual(locality, beforeLocality) {
		t.Fatal("delta evidence mutated an input")
	}
	if evidence.EvidenceSpans.Records[0].EvidenceID == changes.Added[0].EvidenceID {
		t.Fatal("evidence records share mutable state with input")
	}
}
