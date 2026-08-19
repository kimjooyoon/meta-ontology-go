package semanticdelta

import (
	"errors"
	"reflect"
	"testing"
)

func TestDiffSnapshotsClassifiesRelationReplacementAsMeaningChanging(t *testing.T) {
	before := Snapshot{
		Nodes: []Node{{ID: "billing://activity/pay", Kind: "Activity"}, {ID: "billing://entity/order", Kind: "Entity"}, {ID: "billing://entity/receipt", Kind: "Entity"}},
		Facts: []Fact{{Subject: "billing://activity/pay", Predicate: "used", Object: "billing://entity/order"}},
	}
	after := Snapshot{
		Nodes: before.Nodes,
		Facts: []Fact{{Subject: "billing://activity/pay", Predicate: "used", Object: "billing://entity/receipt"}},
	}

	delta, err := DiffSnapshots(before, after)
	if err != nil {
		t.Fatal(err)
	}
	want := Delta{
		AddedFacts:   []Fact{{Subject: "billing://activity/pay", Predicate: "used", Object: "billing://entity/receipt"}},
		RemovedFacts: []Fact{{Subject: "billing://activity/pay", Predicate: "used", Object: "billing://entity/order"}},
	}
	if !reflect.DeepEqual(delta, want) {
		t.Fatalf("relation replacement delta = %#v, want %#v", delta, want)
	}
	if delta.IsEmpty() {
		t.Fatal("relation replacement was classified as syntax-only")
	}
}
func TestDiffSnapshotsDoesNotAliasOpaqueFactValues(t *testing.T) {
	before := Snapshot{Facts: []Fact{{Subject: "a", Predicate: "b\x00c", Object: "d"}}}
	after := Snapshot{Facts: []Fact{{Subject: "a\x00b", Predicate: "c", Object: "d"}}}

	delta, err := DiffSnapshots(before, after)
	if err != nil {
		t.Fatal(err)
	}
	if len(delta.AddedFacts) != 1 || len(delta.RemovedFacts) != 1 {
		t.Fatalf("opaque fact values were aliased: %#v", delta)
	}
}
func TestAdapterApplyRejectsOutOfScopeWithoutCommit(t *testing.T) {
	before := fakeIR{nodes: []Node{{ID: "billing://activity/pay", Kind: "Activity"}}}
	after := fakeIR{nodes: []Node{
		{ID: "billing://activity/pay", Kind: "Activity"},
		{ID: "fraud://entity/charge", Kind: "Entity"},
	}}
	adapter := Adapter[fakeIR]{
		Nodes: func(value fakeIR) ([]Node, error) { return value.nodes, nil },
		Facts: func(value fakeIR) ([]Fact, error) { return value.facts, nil },
	}
	commits := 0
	_, err := adapter.Apply(before, after, Scope{Prefixes: []string{"billing://"}}, func(Delta) error {
		commits++
		return nil
	})
	var scopeErr *ScopeError
	if !errors.As(err, &scopeErr) {
		t.Fatalf("Apply error = %v, want ScopeError", err)
	}
	if commits != 0 {
		t.Fatalf("out-of-scope delta reached commit callback %d time(s)", commits)
	}
}
