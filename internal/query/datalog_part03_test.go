package query

import (
	"testing"
)

func TestDatalogCandidatesAreVisibleButNeverRuleInputs(t *testing.T) {
	graph := New()
	assertAdd(t, graph, NewFact(id("urn:entity:a"), WasDerivedFrom, id("urn:entity:b")))
	assertAdd(t, graph, NewCandidateFact(id("urn:entity:b"), WasDerivedFrom, id("urn:entity:c"), "unresolved"))
	rule := Rule{
		ID:   "direct/v1",
		Head: Triple("dependsOn", Variable("subject"), Variable("source")),
		Body: []Atom{Triple("wasDerivedFrom", Variable("subject"), Variable("source"))},
	}
	derived, err := graph.EvaluateDatalog(DatalogQuery{
		Patterns: []Atom{Triple("dependsOn", Variable("subject"), Variable("source"))},
		Rules:    []Rule{rule}, IncludeDerived: true, IncludeCandidates: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(derived.Derived) != 1 || len(derived.Rows) != 1 {
		t.Fatalf("candidate fact entered rule closure: %#v", derived)
	}
	if derived.Derived[0].Origin != DatalogDerived || derived.Derived[0].Subject != id("urn:entity:a") {
		t.Fatalf("derived origin = %#v", derived.Derived[0])
	}
	candidates, err := graph.EvaluateDatalog(DatalogQuery{
		Patterns:          []Atom{Triple("wasDerivedFrom", Variable("subject"), Variable("source"))},
		IncludeCandidates: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates.Rows) != 2 || candidates.Rows[0].Facts[0].Origin != DatalogDeclared ||
		candidates.Rows[1].Facts[0].Origin != DatalogCandidate {
		t.Fatalf("candidate visibility/order = %#v", candidates)
	}
}
