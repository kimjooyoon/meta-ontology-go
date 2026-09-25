package semanticdelta

import (
	"reflect"
	"testing"
)

func TestReconcileRejectsFactWithMissingEndpoint(t *testing.T) {
	before := Snapshot{Nodes: []Node{{ID: "a", Kind: "Activity"}}}
	delta := Delta{AddedFacts: []Fact{{Subject: "a", Predicate: "uses", Object: "missing"}}}
	originalBefore := before
	originalDelta := delta

	if _, err := Reconcile(before, delta); err == nil {
		t.Fatal("expected a fact with a missing endpoint to fail")
	}
	if !reflect.DeepEqual(before, originalBefore) {
		t.Fatalf("failed reconciliation mutated before snapshot: got %+v", before)
	}
	if !reflect.DeepEqual(delta, originalDelta) {
		t.Fatalf("failed reconciliation mutated delta: got %+v", delta)
	}
}
func TestReconcileAllowsExplicitIncidentFactRemoval(t *testing.T) {
	before := Snapshot{
		Nodes: []Node{{ID: "a", Kind: "Activity"}, {ID: "b", Kind: "Order"}},
		Facts: []Fact{{Subject: "a", Predicate: "uses", Object: "b"}},
	}
	delta := Delta{
		RemovedNodes: []Node{{ID: "a", Kind: "Activity"}},
		RemovedFacts: []Fact{{Subject: "a", Predicate: "uses", Object: "b"}},
	}

	got, err := Reconcile(before, delta)
	if err != nil {
		t.Fatalf("explicit incident fact removal failed: %v", err)
	}
	want := Snapshot{Nodes: []Node{{ID: "b", Kind: "Order"}}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("explicit incident fact removal produced %+v, want %+v", got, want)
	}
}
