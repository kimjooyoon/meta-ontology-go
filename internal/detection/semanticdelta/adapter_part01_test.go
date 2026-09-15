package semanticdelta

import (
	"errors"
	"reflect"
	"testing"
)

type fakeIR struct {
	nodes []Node
	facts []Fact
}

func TestAdapterDiffUsesOnlyConfiguredCallbacks(t *testing.T) {
	before := fakeIR{
		nodes: []Node{{ID: "billing://activity/pay", Kind: "Activity"}},
		facts: []Fact{{Subject: "billing://activity/pay", Predicate: "prov:used", Object: "billing://entity/order"}},
	}
	after := fakeIR{
		nodes: []Node{{ID: "billing://activity/pay", Kind: "Activity"}},
		facts: append(append([]Fact(nil), before.facts...), Fact{Subject: "billing://activity/pay", Predicate: "gooo:invokes", Object: "fraud://activity/check"}),
	}
	adapter := Adapter[fakeIR]{
		Nodes: func(value fakeIR) ([]Node, error) { return value.nodes, nil },
		Facts: func(value fakeIR) ([]Fact, error) { return value.facts, nil },
	}
	delta, err := adapter.Diff(before, after)
	if err != nil {
		t.Fatal(err)
	}
	want := Delta{AddedFacts: []Fact{{Subject: "billing://activity/pay", Predicate: "gooo:invokes", Object: "fraud://activity/check"}}}
	if !reflect.DeepEqual(delta, want) {
		t.Fatalf("delta = %#v, want %#v", delta, want)
	}
}
func TestAdapterRequiresBothCallbacksAndWrapsErrors(t *testing.T) {
	_, err := (Adapter[fakeIR]{}).Snapshot(fakeIR{})
	if err == nil {
		t.Fatal("empty adapter was accepted")
	}
	want := errors.New("source unavailable")
	adapter := Adapter[fakeIR]{
		Nodes: func(fakeIR) ([]Node, error) { return nil, want },
		Facts: func(fakeIR) ([]Fact, error) { return nil, nil },
	}
	_, err = adapter.Snapshot(fakeIR{})
	if !errors.Is(err, want) {
		t.Fatalf("adapter error = %v, want wrapped source error", err)
	}
}
func TestDiffSnapshotsIgnoresOrderingAndPresentationBoundary(t *testing.T) {
	before := Snapshot{
		Nodes: []Node{{ID: "b", Kind: "Entity"}, {ID: "a", Kind: "Activity"}},
		Facts: []Fact{{Subject: "a", Predicate: "uses", Object: "b"}},
	}
	after := Snapshot{
		Nodes: []Node{{ID: "a", Kind: "Activity"}, {ID: "b", Kind: "Entity"}, {ID: "c", Kind: "Entity"}},
		Facts: []Fact{{Subject: "a", Predicate: "uses", Object: "b"}, {Subject: "a", Predicate: "invokes", Object: "c"}},
	}
	delta, err := DiffSnapshots(before, after)
	if err != nil {
		t.Fatal(err)
	}
	if len(delta.AddedNodes) != 1 || len(delta.AddedFacts) != 1 || !reflect.DeepEqual(delta.RemovedNodes, []Node(nil)) {
		t.Fatalf("unexpected snapshot delta: %#v", delta)
	}
	if delta.AddedNodes[0].ID != "c" || delta.AddedFacts[0].Object != "c" {
		t.Fatalf("delta ordering/content = %#v", delta)
	}
}
