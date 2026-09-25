package query

import (
	"errors"
	"reflect"
	"testing"
)

func TestDerivedCandidateClosureAndLimit(t *testing.T) {
	root := id("urn:derived:candidate:root")
	first := id("urn:derived:candidate:first")
	second := id("urn:derived:candidate:second")
	third := id("urn:derived:candidate:third")
	graph := New()
	assertAdd(t, graph, NewFact(root, WasDerivedFrom, first))
	assertAdd(t, graph, NewCandidateFact(root, WasDerivedFrom, second, "ambiguous"))
	assertAdd(t, graph, NewFact(root, WasDerivedFrom, third))
	response, err := graph.Execute(derivedEnvelope(root, RuleDependsOn, LayerAll, 1, 1))
	if err != nil {
		t.Fatal(err)
	}
	full, err := graph.Execute(derivedEnvelope(root, RuleDependsOn, LayerAll, 1, 10))
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Result.DerivedDeterministic) != 1 || len(response.Result.DerivedCandidates) != 0 {
		t.Fatalf("limit did not prefer deterministic derived rows: %#v", response.Result)
	}
	if len(full.Result.DerivedDeterministic) < 1 ||
		!reflect.DeepEqual(response.Result.DerivedDeterministic, full.Result.DerivedDeterministic[:1]) {
		t.Fatalf("limit changed canonical deterministic prefix: %#v vs %#v", response.Result, full.Result)
	}
	candidates, err := graph.Execute(derivedEnvelope(root, RuleDependsOn, LayerCandidate, 1, 10))
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates.Result.DerivedDeterministic) != 0 || len(candidates.Result.DerivedCandidates) != 1 {
		t.Fatalf("candidate layer leaked: %#v", candidates.Result)
	}
	if candidates.Result.DerivedCandidates[0].SourceLayer != FactCandidate.String() {
		t.Fatalf("candidate source layer was promoted: %#v", candidates.Result.DerivedCandidates[0])
	}
	direct, err := graph.Derive(root, DerivedOptions{
		Rule: RuleDependsOn, MaxDepth: 1, Limit: 1, Selection: SelectAll,
	})
	if err != nil || len(direct.Deterministic) != 1 || len(direct.Candidates) != 0 {
		t.Fatalf("direct derived API did not enforce row limit: %#v %v", direct, err)
	}
}
func TestDerivedBoundsRejectInvalidDirectOptions(t *testing.T) {
	graph := New()
	root, target := id("urn:derived:bounds:root"), id("urn:derived:bounds:target")
	assertAdd(t, graph, NewFact(root, WasDerivedFrom, target))
	cases := []DerivedOptions{
		{Rule: RuleDependsOn, MaxDepth: 0, Limit: 1, Selection: SelectAll},
		{Rule: RuleDependsOn, MaxDepth: MaxEnvelopeDepth + 1, Limit: 1, Selection: SelectAll},
		{Rule: RuleDependsOn, MaxDepth: 1, Limit: 0, Selection: SelectAll},
		{Rule: RuleDependsOn, MaxDepth: 1, Limit: MaxEnvelopeLimit + 1, Selection: SelectAll},
	}
	for _, options := range cases {
		if _, err := graph.Derive(root, options); !errors.Is(err, ErrInvalidDerivedQuery) {
			t.Fatalf("invalid direct options returned %v", err)
		}
	}
}
