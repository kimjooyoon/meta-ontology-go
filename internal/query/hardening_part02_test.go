package query

import (
	"errors"
	"reflect"
	"testing"
)

func TestUnknownIDsAndRelationsFailClosedWithoutMutation(t *testing.T) {
	graph := New()
	assertAdd(t, graph, NewFact(id("billing://activity/pay"), Used, id("billing://entity/order")))
	beforeCanonical := graph.Canonical()
	beforeNodes := graph.Nodes()
	missing := NewExactQuery(id("billing://activity/missing"), Used, id("billing://entity/order"))
	if _, err := graph.ExactMatch(missing); !errors.Is(err, ErrUnknownEndpoint) {
		t.Fatalf("unknown exact endpoint error = %v", err)
	}
	if _, err := graph.Traverse(id("billing://activity/missing"), TraversalOptions{MaxDepth: 1}); !errors.Is(err, ErrUnknownEndpoint) {
		t.Fatalf("unknown traversal endpoint error = %v", err)
	}
	unknownRelation := NewExactQuery(id("billing://activity/pay"), Relation("unknown"), id("billing://entity/order"))
	if _, err := graph.ExactMatch(unknownRelation); !errors.Is(err, ErrInvalidQuery) {
		t.Fatalf("unknown relation error = %v", err)
	}
	if graph.Canonical() != beforeCanonical || !reflect.DeepEqual(graph.Nodes(), beforeNodes) {
		t.Fatal("failed query mutated the graph")
	}
}
func TestFactLayerFilteringAndQueryMetadata(t *testing.T) {
	graph := New()
	activity := id("billing://activity/pay")
	order := id("billing://entity/order")
	invoice := id("billing://entity/invoice")
	assertAdd(t, graph, NewFact(activity, Used, order))
	assertAdd(t, graph, NewCandidateFact(activity, Used, invoice, "unresolved"))

	deterministic, err := graph.ExactMatchFiltered(NewExactQuery(activity, Used, order), SelectDeterministic)
	if err != nil || len(deterministic.Deterministic) != 1 || len(deterministic.Candidates) != 0 {
		t.Fatalf("deterministic filter failed: %#v %v", deterministic, err)
	}
	candidate, err := graph.ExactMatchFiltered(NewExactQuery(activity, Used, invoice), SelectCandidate)
	if err != nil || len(candidate.Deterministic) != 0 || len(candidate.Candidates) != 1 {
		t.Fatalf("candidate filter failed: %#v %v", candidate, err)
	}
	all, err := graph.Traverse(activity, TraversalOptions{MaxDepth: 1, Selection: SelectAll})
	if err != nil || len(all.Deterministic) != 1 || len(all.Candidates) != 1 {
		t.Fatalf("all-layer traversal failed: %#v %v", all, err)
	}
	filtered, err := graph.Traverse(activity, TraversalOptions{MaxDepth: 1, Selection: SelectDeterministic})
	if err != nil || len(filtered.Deterministic) != 1 || len(filtered.Candidates) != 0 {
		t.Fatalf("deterministic traversal filter failed: %#v %v", filtered, err)
	}
	candidatePaths, err := graph.Traverse(activity, TraversalOptions{MaxDepth: 1, Selection: SelectCandidate})
	if err != nil || len(candidatePaths.Deterministic) != 0 || len(candidatePaths.Candidates) != 1 {
		t.Fatalf("candidate traversal filter failed: %#v %v", candidatePaths, err)
	}
	if deterministic.Metadata.ProjectionStatus != "unbound" || candidate.Metadata.GraphHash != graph.StableHash() {
		t.Fatalf("query metadata did not describe derived view: %#v %#v", deterministic.Metadata, candidate.Metadata)
	}
}
func TestTypedNodesRejectMismatchedRelations(t *testing.T) {
	graph := New()
	if err := graph.AddNode(Node{ID: id("billing://entity/order"), Kind: EntityNodeKind}); err != nil {
		t.Fatal(err)
	}
	if err := graph.AddNode(Node{ID: id("billing://entity/payment"), Kind: EntityNodeKind}); err != nil {
		t.Fatal(err)
	}
	if err := graph.Add(NewFact(id("billing://entity/order"), Used, id("billing://entity/payment"))); err == nil {
		t.Fatal("typed relation mismatch was accepted")
	}
}
