package semanticdelta

import (
	"reflect"
	"testing"
)

func TestDiffReconcileRoundTripIsCanonical(t *testing.T) {
	before := Snapshot{
		Nodes: []Node{{ID: "billing://entity/order", Kind: "Entity"}, {ID: "billing://activity/pay", Kind: "Activity"}},
		Facts: []Fact{{Subject: "billing://activity/pay", Predicate: "prov:used", Object: "billing://entity/order"}},
	}
	after := Snapshot{
		Nodes: []Node{{ID: "billing://entity/order", Kind: "Entity"}, {ID: "billing://entity/receipt", Kind: "Entity"}},
		Facts: []Fact{{Subject: "billing://entity/receipt", Predicate: "prov:wasGeneratedBy", Object: "billing://entity/order"}},
	}
	delta, err := DiffSnapshots(before, after)
	if err != nil {
		t.Fatal(err)
	}
	reconciled, err := Reconcile(before, delta)
	if err != nil {
		t.Fatal(err)
	}
	want, err := after.Normalized()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(reconciled, want) {
		t.Fatalf("reconciled snapshot = %#v, want %#v", reconciled, want)
	}
}

func TestReconcileRejectsTamperedDeltaWithoutMutation(t *testing.T) {
	before := Snapshot{
		Nodes: []Node{{ID: "billing://entity/order", Kind: "Entity"}},
		Facts: []Fact{{Subject: "billing://entity/order", Predicate: "prov:used", Object: "billing://entity/customer"}},
	}
	delta := Delta{RemovedFacts: []Fact{{Subject: "billing://entity/order", Predicate: "prov:used", Object: "billing://entity/missing"}}}
	originalBefore := before
	originalDelta := delta
	if _, err := Reconcile(before, delta); err == nil {
		t.Fatal("tampered removal was accepted")
	}
	if !reflect.DeepEqual(before, originalBefore) || !reflect.DeepEqual(delta, originalDelta) {
		t.Fatalf("reconciliation mutated inputs: before=%#v delta=%#v", before, delta)
	}
}

func TestReconcileIsInvariantToInputPermutation(t *testing.T) {
	before := Snapshot{
		Nodes: []Node{{ID: "b", Kind: "Entity"}, {ID: "a", Kind: "Activity"}},
		Facts: []Fact{{Subject: "a", Predicate: "uses", Object: "b"}},
	}
	delta := Delta{
		AddedNodes:   []Node{{ID: "d", Kind: "Entity"}, {ID: "c", Kind: "Entity"}},
		RemovedNodes: []Node{{ID: "a", Kind: "Activity"}},
		AddedFacts:   []Fact{{Subject: "c", Predicate: "uses", Object: "d"}, {Subject: "d", Predicate: "uses", Object: "b"}},
		RemovedFacts: []Fact{{Subject: "a", Predicate: "uses", Object: "b"}},
	}
	permuted := Snapshot{
		Nodes: []Node{{ID: "a", Kind: "Activity"}, {ID: "b", Kind: "Entity"}},
		Facts: []Fact{{Subject: "a", Predicate: "uses", Object: "b"}},
	}
	permutedDelta := Delta{
		AddedNodes:   []Node{{ID: "c", Kind: "Entity"}, {ID: "d", Kind: "Entity"}},
		RemovedNodes: []Node{{ID: "a", Kind: "Activity"}},
		AddedFacts:   []Fact{{Subject: "d", Predicate: "uses", Object: "b"}, {Subject: "c", Predicate: "uses", Object: "d"}},
		RemovedFacts: []Fact{{Subject: "a", Predicate: "uses", Object: "b"}},
	}
	first, err := Reconcile(before, delta)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Reconcile(permuted, permutedDelta)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("permuted reconciliation changed meaning: %#v != %#v", first, second)
	}
}

func TestReconcileRejectsDanglingFactAfterNodeRemoval(t *testing.T) {
	before := Snapshot{
		Nodes: []Node{{ID: "a", Kind: "Activity"}, {ID: "b", Kind: "Order"}},
		Facts: []Fact{{Subject: "a", Predicate: "uses", Object: "b"}},
	}
	delta := Delta{RemovedNodes: []Node{{ID: "a", Kind: "Activity"}}}
	originalBefore := before
	originalDelta := delta

	if _, err := Reconcile(before, delta); err == nil {
		t.Fatal("expected node removal with a retained incident fact to fail")
	}
	if !reflect.DeepEqual(before, originalBefore) {
		t.Fatalf("failed reconciliation mutated before snapshot: got %+v", before)
	}
	if !reflect.DeepEqual(delta, originalDelta) {
		t.Fatalf("failed reconciliation mutated delta: got %+v", delta)
	}
}

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
