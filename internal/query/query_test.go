package query

import (
	"reflect"
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
		if err != nil {
			t.Fatal(err)
		}
		if !result.Empty() {
			t.Fatalf("non-exact query matched: %#v", result)
		}
	}
}

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
	if _, err := graph.Traverse(id("billing://activity/pay"), TraversalOptions{}); err == nil {
		t.Fatal("unbounded traversal was accepted")
	}
}

func id(raw string) ID {
	parsed, err := ParseID(raw)
	if err != nil {
		panic(err)
	}
	return parsed
}

func assertAdd(t *testing.T, graph *Graph, fact Fact) {
	t.Helper()
	if err := graph.Add(fact); err != nil {
		t.Fatal(err)
	}
}

func pathIDs(paths []Path) [][]ID {
	result := make([][]ID, len(paths))
	for index, path := range paths {
		result[index] = append([]ID(nil), path.IDs...)
	}
	return result
}
