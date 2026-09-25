package bidir

import (
	"testing"
)

func TestBillingGoooDogfoodRoundTripsDSLGoAndBack(t *testing.T) {
	document := billingGoooDocument(t)
	base, err := Get(document)
	if err != nil {
		t.Fatal(err)
	}
	if err := CheckGetPut(document); err != nil {
		t.Fatalf("Get-Put: %v", err)
	}

	projected := ProjectFacts(base)
	for _, fact := range projected {
		if !fact.Source.Valid() {
			t.Fatalf("projected fact lost source provenance: %#v", fact)
		}
	}
	lifted, err := LiftFacts(base, projected)
	if err != nil {
		t.Fatalf("project/lift: %v", err)
	}
	if !SemanticEquivalent(base, lifted.Model) {
		t.Fatalf("project/lift changed meaning: %s != %s", SemanticFingerprint(base), SemanticFingerprint(lifted.Model))
	}

	derived := NewSourcedFact(
		DeterministicFact,
		"billing://entity/payment",
		PredicateWasDerivedFrom,
		"billing://entity/order",
		SourceSpan{File: "examples/billing/handwritten.go", Start: 6, End: 24},
	)
	derived.SubjectKind = EntityKind
	derived.ObjectKind = EntityKind
	updated, err := Reconcile(base, FactDelta{Added: FactSet{derived}})
	if err != nil {
		t.Fatalf("lift Go fact: %v", err)
	}
	if len(updated.Delta.AddedRelations) != 1 || !updated.Locality.Contains(derived.Subject) || !updated.Locality.Contains(derived.Object) {
		t.Fatalf("Go fact delta was not localized: delta=%#v locality=%#v", updated.Delta, updated.Locality)
	}

	written, err := Put(document, updated.Model)
	if err != nil {
		t.Fatalf("Put accepted Go fact: %v", err)
	}
	observed, err := Get(written)
	if err != nil {
		t.Fatalf("Get written document: %v", err)
	}
	if !SemanticEquivalent(updated.Model, observed) {
		t.Fatalf("DSL -> Go -> DSL changed meaning: %s != %s", SemanticFingerprint(updated.Model), SemanticFingerprint(observed))
	}
	relation, found := findRelation(observed, PredicateWasDerivedFrom, derived.Subject, derived.Object)
	if !found || relation.Span != derived.Source {
		t.Fatalf("accepted provenance was not preserved: found=%v relation=%#v", found, relation)
	}
}
