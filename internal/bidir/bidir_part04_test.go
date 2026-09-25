package bidir

import (
	"reflect"
	"testing"
)

func TestDiffApplyPreservesSemanticDeltaAndIgnoresPresentation(t *testing.T) {
	base, err := Get(billingDocument())
	if err != nil {
		t.Fatal(err)
	}
	updated := base.Clone()
	updated.Nodes[0].Name = "Display Rename"
	updated.Relations = append(updated.Relations, Relation{
		Kind:   PredicateInvokes,
		Source: "billing://activity/pay-order",
		Target: "billing://activity/audit-payment",
	})
	delta := Diff(base, updated)
	if len(delta.AddedRelations) != 1 || len(delta.AddedNodes) != 0 || len(delta.RemovedNodes) != 0 {
		t.Fatalf("unexpected semantic delta: %#v", delta)
	}
	applied, err := base.Apply(delta)
	if err != nil {
		t.Fatal(err)
	}
	if !SemanticEquivalent(updated, applied) {
		t.Fatalf("delta application changed meaning: %s != %s", SemanticFingerprint(updated), SemanticFingerprint(applied))
	}
	presentationOnly := base.Clone()
	presentationOnly.Nodes[0].Name = "Another Rename"
	if delta := Diff(base, presentationOnly); !delta.IsEmpty() {
		t.Fatalf("presentation-only edit produced semantic delta: %#v", delta)
	}
}
func TestFactSetNormalizationIsDeterministic(t *testing.T) {
	one := NewSourcedFact(DeterministicFact, "b", PredicateInvokes, "c", SourceSpan{File: "b.go", Start: 1, End: 2})
	two := NewSourcedFact(DeterministicFact, "a", PredicateInvokes, "c", SourceSpan{File: "a.go", Start: 1, End: 2})
	set := FactSet{one, two, one}
	want := FactSet{two, one}
	if got := set.Normalized(); !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected normalized facts:\ngot  %#v\nwant %#v", got, want)
	}
}
