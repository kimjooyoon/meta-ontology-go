package semanticdelta

import (
	"reflect"
	"testing"
)

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
