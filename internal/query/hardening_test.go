package query

import (
	"errors"
	"reflect"
	"testing"
)

func TestPermutationAndRepeatProduceIdenticalReadResults(t *testing.T) {
	facts := []Fact{
		NewFact(id("billing://activity/pay"), Used, id("billing://entity/order")),
		NewFact(id("billing://entity/payment"), WasGeneratedBy, id("billing://activity/pay")),
		NewCandidateFact(id("billing://entity/order"), WasDerivedFrom, id("billing://entity/archive"), "ambiguous"),
	}
	first := New()
	second := New()
	for _, fact := range facts {
		assertAdd(t, first, fact)
	}
	for index := len(facts) - 1; index >= 0; index-- {
		assertAdd(t, second, facts[index])
	}
	if first.Canonical() != second.Canonical() || first.StableHash() != second.StableHash() {
		t.Fatalf("permuted graph changed canonical projection")
	}
	if !reflect.DeepEqual(first.Nodes(), second.Nodes()) || !reflect.DeepEqual(first.Relations(), second.Relations()) {
		t.Fatalf("permuted graph changed node or relation order")
	}
	want, err := first.Traverse(id("billing://activity/pay"), TraversalOptions{MaxDepth: 2})
	if err != nil {
		t.Fatal(err)
	}
	for run := 0; run < 5; run++ {
		got, err := second.Traverse(id("billing://activity/pay"), TraversalOptions{MaxDepth: 2})
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("repeat %d changed traversal result", run)
		}
	}
}

func TestCycleTraversalIsSimpleAndBounded(t *testing.T) {
	graph := New()
	assertAdd(t, graph, NewFact(id("urn:cycle:a"), WasDerivedFrom, id("urn:cycle:b")))
	assertAdd(t, graph, NewFact(id("urn:cycle:b"), WasDerivedFrom, id("urn:cycle:c")))
	assertAdd(t, graph, NewFact(id("urn:cycle:c"), WasDerivedFrom, id("urn:cycle:a")))
	result, err := graph.Traverse(id("urn:cycle:a"), TraversalOptions{MaxDepth: 8})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Deterministic) != 2 {
		t.Fatalf("cycle emitted unexpected paths: %#v", result.Deterministic)
	}
	for _, path := range result.Deterministic {
		if path.Depth() > 8 || hasRepeatedID(path.IDs) {
			t.Fatalf("cycle path was not simple/bounded: %#v", path)
		}
	}
}

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

func hasRepeatedID(ids []ID) bool {
	seen := make(map[ID]struct{}, len(ids))
	for _, id := range ids {
		if _, exists := seen[id]; exists {
			return true
		}
		seen[id] = struct{}{}
	}
	return false
}
