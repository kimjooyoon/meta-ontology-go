package query

import (
	"reflect"
	"testing"
)

func TestBoundedTraversalIsDeterministicAndCandidateAware(t *testing.T) {
	graph := New()
	activity := id("billing://activity/pay")
	order := id("billing://entity/order")
	invoice := id("billing://entity/invoice")
	archive := id("billing://entity/archive")
	assertAdd(t, graph, NewFact(order, WasDerivedFrom, archive))
	assertAdd(t, graph, NewCandidateFact(order, Used, invoice, "ambiguous helper"))
	assertAdd(t, graph, NewFact(activity, Used, order))

	result, err := graph.Traverse(activity, TraversalOptions{Predicate: Used, MaxDepth: 2})
	if err != nil {
		t.Fatal(err)
	}
	wantDeterministic := [][]ID{{activity, order}}
	if !reflect.DeepEqual(pathIDs(result.Deterministic), wantDeterministic) {
		t.Fatalf("unexpected deterministic paths: got %#v want %#v", pathIDs(result.Deterministic), wantDeterministic)
	}
	wantCandidates := [][]ID{{activity, order, invoice}}
	if !reflect.DeepEqual(pathIDs(result.Candidates), wantCandidates) {
		t.Fatalf("unexpected candidate paths: got %#v want %#v", pathIDs(result.Candidates), wantCandidates)
	}
	if result.Candidates[0].Status != Candidate {
		t.Fatalf("candidate path lost status: %#v", result.Candidates[0])
	}

	short, err := graph.Traverse(activity, TraversalOptions{MaxDepth: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(short.Deterministic) != 1 || len(short.Candidates) != 0 {
		t.Fatalf("bounded traversal crossed depth limit: %#v", short)
	}
}
