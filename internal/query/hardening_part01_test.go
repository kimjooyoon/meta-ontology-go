package query

import (
	"reflect"
	"slices"
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
	for _, fact := range slices.Backward(facts) {
		assertAdd(t, second, fact)
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
	for run := range 5 {
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
