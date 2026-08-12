package query

import (
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
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

func TestFromSemanticIRKeepsCandidatesOutOfAuthoritativeQueries(t *testing.T) {
	ir := semantic.NewIR("billing", "billing")
	activity, err := semantic.NewActivity("billing://activity/pay", "billing", "PayOrder")
	if err != nil {
		t.Fatal(err)
	}
	order, err := semantic.NewEntity("billing://entity/order", "billing", "Order")
	if err != nil {
		t.Fatal(err)
	}
	if err := ir.AddNode(activity); err != nil {
		t.Fatal(err)
	}
	if err := ir.AddNode(order); err != nil {
		t.Fatal(err)
	}
	invoice, err := semantic.NewEntity("billing://entity/invoice", "billing", "Invoice")
	if err != nil {
		t.Fatal(err)
	}
	if err := ir.AddNode(invoice); err != nil {
		t.Fatal(err)
	}
	if err := ir.AddFact(semantic.NewUsedFact(activity.ID, order.ID)); err != nil {
		t.Fatal(err)
	}
	if err := ir.AddCandidate(semantic.NewCandidateFact(activity.ID, semantic.Used, order.ID, "ambiguous adapter")); err != nil {
		t.Fatal(err)
	}
	if err := ir.AddCandidate(semantic.NewCandidateFact(activity.ID, semantic.Used, invoice.ID, "ambiguous invoice adapter")); err != nil {
		t.Fatal(err)
	}
	projected, err := FromSemanticIR(ir)
	if err != nil {
		t.Fatal(err)
	}
	result, err := projected.ExactMatch(NewExactQuery(ID(activity.ID.String()), Used, ID(order.ID.String())))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Deterministic) != 1 || len(result.Candidates) != 0 {
		t.Fatalf("candidate was not shadowed by authoritative fact: %#v", result)
	}
	candidateResult, err := projected.ExactMatch(NewExactQuery(ID(activity.ID.String()), Used, ID(invoice.ID.String())))
	if err != nil {
		t.Fatal(err)
	}
	if len(candidateResult.Deterministic) != 0 || len(candidateResult.Candidates) != 1 || candidateResult.Candidates[0].Status != Candidate {
		t.Fatalf("candidate projection was not kept separate: %#v", candidateResult)
	}
	if projected.StableHash() == "" || projected.Canonical() == "" {
		t.Fatal("query projection did not expose a stable read fingerprint")
	}
}

func TestFromSemanticIRRejectsInvalidAuthoritativeGraph(t *testing.T) {
	ir := semantic.NewIR("billing", "billing")
	activity, err := semantic.NewActivity("billing://activity/pay", "billing", "PayOrder")
	if err != nil {
		t.Fatal(err)
	}
	entity, err := semantic.NewEntity("billing://entity/order", "billing", "Order")
	if err != nil {
		t.Fatal(err)
	}
	if err := ir.AddNode(activity); err != nil {
		t.Fatal(err)
	}
	if err := ir.AddNode(entity); err != nil {
		t.Fatal(err)
	}
	// Reverse the PROV direction deliberately. The adapter must validate the
	// IR first rather than silently publishing a query projection.
	if err := ir.AddFact(semantic.NewFact(activity.ID, semantic.WasGeneratedBy, entity.ID)); err != nil {
		// AddFact currently permits construction and IR.Validate owns the final
		// graph boundary; retain the test for either fail-closed behavior.
		return
	}
	if _, err := FromSemanticIR(ir); err == nil {
		t.Fatal("invalid semantic relation was projected into query graph")
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
