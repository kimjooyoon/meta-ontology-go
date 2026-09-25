package bidir

import (
	"reflect"
	"testing"
)

func TestUnknownEndpointRejectedWithoutImplicitNodeOrMutation(t *testing.T) {
	base, err := Get(billingDocument())
	if err != nil {
		t.Fatal(err)
	}
	original := base.Clone()
	fact := NewSourcedFact(DeterministicFact, "billing://activity/pay-order", PredicateInvokes, "billing://activity/unregistered", SourceSpan{File: "lift.go", Start: 1, End: 2})
	fact.SubjectKind = ActivityKind
	fact.ObjectKind = ActivityKind
	result, err := Reconcile(base, FactDelta{Added: FactSet{fact}})
	if !hasConflict(err, ConflictUnknownEndpoint) {
		t.Fatalf("unknown endpoint was accepted: %v", err)
	}
	if !reflect.DeepEqual(base, original) || !reflect.DeepEqual(result.Model, original) {
		t.Fatalf("unknown endpoint mutated model: base=%#v result=%#v", base, result.Model)
	}
	if _, exists := result.Model.node(fact.Object); exists {
		t.Fatal("unknown endpoint was implicitly registered")
	}
}
func TestUnknownEndpointPartialObservationPreservesNoDeleteAndNoWrite(t *testing.T) {
	base, err := Get(billingDocument())
	if err != nil {
		t.Fatal(err)
	}
	original := base.Clone()
	partial := NewSourcedFact(SyntacticFact, "billing://activity/pay-order", PredicateInvokes, "billing://activity/unregistered", SourceSpan{File: "partial.go", Start: 3, End: 4})
	unknown := NewSourcedFact(DeterministicFact, "billing://activity/pay-order", PredicateInvokes, "billing://activity/unregistered", SourceSpan{File: "lift.go", Start: 1, End: 2})
	unknown.SubjectKind = ActivityKind
	unknown.ObjectKind = ActivityKind
	result, err := Reconcile(base, FactDelta{Added: FactSet{partial, unknown}})
	if !hasConflict(err, ConflictUnknownEndpoint) {
		t.Fatalf("unknown endpoint was not rejected: %v", err)
	}
	if len(result.Delta.RemovedRelations) != 0 || len(result.Delta.RemovedNodes) != 0 {
		t.Fatalf("partial observation created removals: %#v", result.Delta)
	}
	if !reflect.DeepEqual(base, original) || !reflect.DeepEqual(result.Model, original) {
		t.Fatalf("rejected partial observation mutated model: base=%#v result=%#v", base, result.Model)
	}
	if _, exists := result.Model.node(unknown.Object); exists {
		t.Fatal("unknown endpoint was implicitly registered")
	}
	if _, exists := findRelation(result.Model, PredicateUsed, "billing://activity/pay-order", "billing://entity/order"); !exists {
		t.Fatal("rejected partial observation deleted an existing relation")
	}
}
func TestPartialObservationDoesNotDeleteAbsentRelations(t *testing.T) {
	base, err := Get(billingDocument())
	if err != nil {
		t.Fatal(err)
	}
	fact := NewSourcedFact(SyntacticFact, "billing://activity/pay-order", PredicateInvokes, "billing://activity/audit-payment", SourceSpan{File: "partial.go", Start: 3, End: 4})
	result, err := Reconcile(base, FactDelta{Added: FactSet{fact}})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Delta.RemovedRelations) != 0 || !result.Delta.IsEmpty() {
		t.Fatalf("partial observation changed absent relations: %#v", result.Delta)
	}
	if !SemanticEquivalent(base, result.Model) {
		t.Fatal("partial observation changed semantic meaning")
	}
}
