package bidir

import (
	"testing"
)

func TestExplicitRemovalIsAFirstClassDelta(t *testing.T) {
	base, err := Get(billingDocument())
	if err != nil {
		t.Fatal(err)
	}
	fact := NewSourcedFact(DeterministicFact, "billing://activity/pay-order", PredicateUsed, "billing://entity/order", SourceSpan{File: "payment.go", Start: 10, End: 20})
	result, err := Reconcile(base, FactDelta{Removed: FactSet{fact}})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Delta.RemovedRelations) != 1 {
		t.Fatalf("expected removed relation: %#v", result.Delta)
	}
	if _, found := findRelation(result.Model, PredicateUsed, fact.Subject, fact.Object); found {
		t.Fatal("removed relation still present")
	}

	second, err := Reconcile(result.Model, FactDelta{Removed: FactSet{NewFact(DeterministicFact, fact.Subject, fact.Predicate, fact.Object)}})
	if err != nil {
		t.Fatal(err)
	}
	if !second.Delta.IsEmpty() {
		t.Fatalf("idempotent removal changed the model: %#v", second.Delta)
	}
}
func TestSemanticEquivalenceIgnoresPresentationAndFingerprintIsStable(t *testing.T) {
	left, err := Get(billingDocument())
	if err != nil {
		t.Fatal(err)
	}
	right := left.Clone()
	right.Nodes[0].Name = "Renamed for display"
	right.Nodes[0].Namespace = "other-display-namespace"
	right.Nodes[0].Aliases = []string{"legacy", "new"}
	right.Nodes[0].Span = SourceSpan{File: "renamed.gooo", Start: 1, End: 8}
	if !SemanticEquivalent(left, right) {
		t.Fatal("presentation-only changes broke semantic equivalence")
	}
	if SemanticFingerprint(left) != SemanticFingerprint(right) {
		t.Fatal("presentation-only changes changed semantic fingerprint")
	}
	right.Nodes[0].Attributes = map[string]string{"effect": "changed"}
	if SemanticEquivalent(left, right) {
		t.Fatal("semantic attribute change was ignored")
	}
}
func TestLocalityExcludesUnrelatedNodes(t *testing.T) {
	base, err := Get(billingDocument())
	if err != nil {
		t.Fatal(err)
	}
	fact := NewSourcedFact(DeterministicFact, "billing://activity/pay-order", PredicateInvokes, "billing://activity/audit-payment", SourceSpan{File: "payment.go", Start: 30, End: 40})
	updated, err := Reconcile(base, FactDelta{Added: FactSet{fact}})
	if err != nil {
		t.Fatal(err)
	}
	locality := updated.Locality
	if !locality.Contains("billing://activity/pay-order") || !locality.Contains("billing://activity/audit-payment") {
		t.Fatalf("changed endpoints absent from locality: %#v", locality)
	}
	if locality.Contains("billing://entity/unrelated") {
		t.Fatalf("unrelated node included in locality: %#v", locality)
	}
	unchanged := LocalityBetween(base, base)
	if len(unchanged.Touched) != 0 || len(unchanged.Affected) != 0 {
		t.Fatalf("implementation-only/no semantic edit should be local to nothing: %#v", unchanged)
	}
}
