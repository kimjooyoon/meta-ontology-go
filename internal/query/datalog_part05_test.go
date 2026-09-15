package query

import (
	"errors"
	"reflect"
	"testing"
)

func TestDatalogResultLimitIsDeterministicAndMarkedIncomplete(t *testing.T) {
	graph := New()
	for _, target := range []ID{id("urn:entity:z"), id("urn:entity:a"), id("urn:entity:m")} {
		assertAdd(t, graph, NewFact(id("urn:activity:start"), Used, target))
	}
	before := graph.Canonical()
	result, err := graph.EvaluateDatalog(DatalogQuery{
		Patterns: []Atom{Triple("used", Constant(id("urn:activity:start")), Variable("entity"))},
		Limit:    2,
	})
	if !errors.Is(err, ErrDatalogBudget) {
		t.Fatalf("row limit error = %v, want ErrDatalogBudget", err)
	}
	if result.Complete || len(result.Rows) != 2 {
		t.Fatalf("bounded result = %#v", result)
	}
	if got, _ := result.Rows[0].Value("entity"); got != id("urn:entity:a") {
		t.Fatalf("first bounded row = %s", got)
	}
	if graph.Canonical() != before {
		t.Fatal("Datalog evaluation mutated graph authority")
	}
	if !reflect.DeepEqual(result.Rows[0].Bindings, map[string]ID{"entity": id("urn:entity:a")}) {
		t.Fatalf("binding escaped canonical shape: %#v", result.Rows[0].Bindings)
	}
}
func TestDatalogExcludedDerivedViewDoesNotConsumeClosureBudget(t *testing.T) {
	graph := New()
	assertAdd(t, graph, NewFact(id("urn:entity:a"), WasDerivedFrom, id("urn:entity:b")))
	assertAdd(t, graph, NewFact(id("urn:entity:b"), WasDerivedFrom, id("urn:entity:c")))
	result, err := graph.EvaluateDatalog(DatalogQuery{
		Patterns: []Atom{Triple("wasDerivedFrom", Variable("subject"), Variable("source"))},
		Rules: []Rule{{
			ID:   "direct/v1",
			Head: Triple("dependsOn", Variable("subject"), Variable("source")),
			Body: []Atom{Triple("wasDerivedFrom", Variable("subject"), Variable("source"))},
		}},
		MaxDerivedFacts: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rows) != 2 || len(result.Derived) != 0 || !result.Complete {
		t.Fatalf("excluded derived view was evaluated: %#v", result)
	}
}
