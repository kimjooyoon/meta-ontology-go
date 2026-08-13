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
