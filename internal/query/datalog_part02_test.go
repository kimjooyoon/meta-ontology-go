package query

import (
	"testing"
)

func TestDatalogConjunctionPermutationHasOneCanonicalResult(t *testing.T) {
	graph := New()
	assertAdd(t, graph, NewFact(id("urn:activity:start"), Used, id("urn:entity:input")))
	assertAdd(t, graph, NewFact(id("urn:activity:start"), WasAssociatedWith, id("urn:agent:operator")))
	rule := Rule{
		ID:   "annotated/v1",
		Head: Triple("annotated", Variable("activity"), Variable("agent")),
		Body: []Atom{
			Triple("wasAssociatedWith", Variable("activity"), Variable("agent")),
			Triple("used", Variable("activity"), Variable("input")),
		},
	}
	forward, err := graph.EvaluateDatalog(DatalogQuery{
		Patterns: []Atom{
			Triple("used", Variable("activity"), Variable("input")),
			Triple("wasAssociatedWith", Variable("activity"), Variable("agent")),
		},
		Rules: []Rule{rule}, IncludeDerived: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	reversed, err := graph.EvaluateDatalog(DatalogQuery{
		Patterns: []Atom{
			Triple("wasAssociatedWith", Variable("activity"), Variable("agent")),
			Triple("used", Variable("activity"), Variable("input")),
		},
		Rules: []Rule{{
			ID:   rule.ID,
			Head: rule.Head,
			Body: []Atom{rule.Body[1], rule.Body[0]},
		}}, IncludeDerived: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if forward.Canonical() != reversed.Canonical() || forward.StableHash() != reversed.StableHash() {
		t.Fatalf(
			"conjunction permutation changed canonical result: %s/%s vs %s/%s",
			forward.Canonical(), forward.StableHash(), reversed.Canonical(), reversed.StableHash(),
		)
	}
	if len(forward.Derived) != 1 || len(forward.Derived[0].Support) != 2 ||
		forward.Derived[0].Support[0].Predicate != string(Used) {
		t.Fatalf("derived support was not normalized: %#v", forward.Derived)
	}
}
