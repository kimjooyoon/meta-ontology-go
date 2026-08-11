package bidir

import (
	"errors"
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

func hasConflict(err error, want ConflictKind) bool {
	var reconcileErr *ReconcileError
	if !errors.As(err, &reconcileErr) {
		return false
	}
	for _, conflict := range reconcileErr.Conflicts {
		if conflict.Kind == want {
			return true
		}
	}
	return false
}
