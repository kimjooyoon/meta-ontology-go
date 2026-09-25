package query

import (
	"errors"
	"reflect"
	"testing"
)

func TestTraversalOrderingAndIncomingDirection(t *testing.T) {
	graph := New()
	start := id("urn:gooo:activity:start")
	alpha := id("urn:gooo:entity/alpha")
	zulu := id("urn:gooo:entity/zulu")
	end := id("urn:gooo:entity/end")
	assertAdd(t, graph, NewFact(start, Used, zulu))
	assertAdd(t, graph, NewFact(start, Used, alpha))
	assertAdd(t, graph, NewFact(zulu, WasDerivedFrom, end))
	assertAdd(t, graph, NewFact(alpha, WasDerivedFrom, end))

	result, err := graph.Traverse(start, TraversalOptions{MaxDepth: 2, Direction: Outgoing})
	if err != nil {
		t.Fatal(err)
	}
	want := [][]ID{
		{start, alpha},
		{start, zulu},
		{start, alpha, end},
		{start, zulu, end},
	}
	if got := pathIDs(result.Deterministic); !reflect.DeepEqual(got, want) {
		t.Fatalf("paths were not deterministically ordered: got %#v want %#v", got, want)
	}

	incoming, err := graph.Traverse(end, TraversalOptions{MaxDepth: 2, Direction: Incoming})
	if err != nil {
		t.Fatal(err)
	}
	wantIncoming := [][]ID{{end, alpha}, {end, zulu}, {end, alpha, start}, {end, zulu, start}}
	if got := pathIDs(incoming.Deterministic); !reflect.DeepEqual(got, wantIncoming) {
		t.Fatalf("incoming paths were not deterministic: got %#v want %#v", got, wantIncoming)
	}
}
func TestInvalidInputsAreRejected(t *testing.T) {
	graph := New()
	if err := graph.Add(NewFact(ID("display-name"), Used, id("billing://entity/order"))); err == nil {
		t.Fatal("non-URI ID was accepted")
	}
	if _, err := graph.ExactMatch(NewExactQuery(id("billing://activity/pay"), Relation("gooo:maybe"), id("billing://entity/order"))); err == nil {
		t.Fatal("unknown relation was accepted")
	}
	if _, err := graph.Traverse(id("billing://activity/pay"), TraversalOptions{}); !errors.Is(err, ErrInvalidTraversal) {
		t.Fatal("unbounded traversal was accepted")
	}
}
