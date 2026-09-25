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
