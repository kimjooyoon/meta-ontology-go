package query

import (
	"errors"
	"testing"
)

func TestExactMatchSeparatesFactStatus(t *testing.T) {
	graph := New()
	assertAdd(t, graph, NewCandidateFact(id("billing://activity/pay"), Used, id("billing://entity/order"), "lifted call"))
	query := NewExactQuery(id("BILLING://ACTIVITY/pay"), Relation("used"), id("billing://entity/order"))
	result, err := graph.ExactMatch(query)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Deterministic) != 0 || len(result.Candidates) != 1 {
		t.Fatalf("unexpected candidate result: %#v", result)
	}

	assertAdd(t, graph, NewFact(id("billing://activity/pay"), Used, id("billing://entity/order")))
	result, err = graph.Match(query)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Deterministic) != 1 || len(result.Candidates) != 0 {
		t.Fatalf("deterministic fact did not shadow candidate: %#v", result)
	}
	if !graph.HasFact(result.Deterministic[0].Key()) || graph.HasCandidate(result.Deterministic[0].Key()) {
		t.Fatal("graph status indexes disagree with exact result")
	}
}
func TestExactMatchIsStrictAboutTripleIdentity(t *testing.T) {
	graph := New()
	assertAdd(t, graph, NewFact(id("billing://activity/pay"), Used, id("billing://entity/order")))
	wrongRelation := NewExactQuery(id("billing://activity/pay"), WasDerivedFrom, id("billing://entity/order"))
	wrongNamespace := NewExactQuery(id("settlement://activity/pay"), Used, id("billing://entity/order"))
	for _, query := range []ExactQuery{wrongRelation, wrongNamespace} {
		result, err := graph.ExactMatch(query)
		if query.Subject == ID("settlement://activity/pay") {
			if !errors.Is(err, ErrUnknownEndpoint) {
				t.Fatalf("unknown endpoint error = %v", err)
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		if !result.Empty() {
			t.Fatalf("non-exact query matched: %#v", result)
		}
	}
}
