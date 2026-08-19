package semanticdelta

import (
	"reflect"
	"testing"
)

func TestAdapterApplyRejectsDanglingFactWithoutCommit(t *testing.T) {
	before := fakeIR{nodes: []Node{{ID: "billing://entity/order", Kind: "Entity"}}}
	after := fakeIR{
		nodes: before.nodes,
		facts: []Fact{{
			Subject: "billing://entity/order", Predicate: "uses", Object: "billing://entity/missing",
		}},
	}
	adapter := Adapter[fakeIR]{
		Nodes: func(value fakeIR) ([]Node, error) { return value.nodes, nil },
		Facts: func(value fakeIR) ([]Fact, error) { return value.facts, nil },
	}
	commits := 0
	_, err := adapter.Apply(before, after, Scope{Prefixes: []string{"billing://"}}, func(Delta) error {
		commits++
		return nil
	})
	if err == nil {
		t.Fatal("Apply accepted a dangling fact endpoint")
	}
	if commits != 0 {
		t.Fatalf("invalid semantic delta reached commit callback %d time(s)", commits)
	}
}
func TestAdapterApplyCommitsAllowedDeltaAndSkipsReplay(t *testing.T) {
	current := fakeIR{nodes: []Node{{ID: "billing://activity/pay", Kind: "Activity"}}}
	desired := fakeIR{nodes: []Node{
		{ID: "billing://activity/pay", Kind: "Activity"},
		{ID: "billing://entity/order", Kind: "Entity"},
	}}
	adapter := Adapter[fakeIR]{
		Nodes: func(value fakeIR) ([]Node, error) { return value.nodes, nil },
		Facts: func(value fakeIR) ([]Fact, error) { return value.facts, nil },
	}
	commits := 0
	var committed Delta
	commit := func(delta Delta) error {
		commits++
		committed = delta
		current = desired
		return nil
	}
	scope := Scope{Prefixes: []string{"billing://"}}
	if report, err := adapter.Apply(current, desired, scope, commit); err != nil || !report.Passes() {
		t.Fatalf("allowed Apply = report %#v, error %v", report, err)
	}
	want := Delta{AddedNodes: []Node{{ID: "billing://entity/order", Kind: "Entity"}}}
	if !reflect.DeepEqual(committed, want) || commits != 1 {
		t.Fatalf("commit = %#v after %d calls, want %#v after one call", committed, commits, want)
	}
	if report, err := adapter.Apply(current, desired, scope, nil); err != nil || !report.Passes() {
		t.Fatalf("replay Apply = report %#v, error %v", report, err)
	}
	if commits != 1 {
		t.Fatalf("replay reached commit callback %d time(s), want one total", commits)
	}
}
