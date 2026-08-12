package query

import "testing"

func TestDerivedQuotaStopsBFSWithoutMutationAndReplays(t *testing.T) {
	root := id("urn:derived:quota:root")
	noise := id("urn:derived:quota:aaa")
	target := id("urn:derived:quota:target")
	first, second := New(), New()
	facts := []Fact{
		NewFact(noise, WasDerivedFrom, target),
		NewFact(root, WasDerivedFrom, target),
	}
	for _, fact := range facts {
		assertAdd(t, first, fact)
	}
	for index := len(facts) - 1; index >= 0; index-- {
		assertAdd(t, second, facts[index])
	}
	beforeCanonical, beforeHash := first.Canonical(), first.StableHash()
	request := derivedEnvelope(root, RuleDependsOn, LayerDeterministic, 1, 1)
	limited, err := first.Execute(request)
	if err != nil {
		t.Fatal(err)
	}
	if len(limited.Result.DerivedDeterministic) != 0 {
		t.Fatalf("edge quota did not stop the bounded BFS: %#v", limited.Result)
	}
	if first.Canonical() != beforeCanonical || first.StableHash() != beforeHash {
		t.Fatal("quota-bounded derived query mutated graph state")
	}
	for run := 0; run < 3; run++ {
		replay, replayErr := first.Execute(request)
		if replayErr != nil || replay.Hash != limited.Hash {
			t.Fatalf("quota-bounded replay %d changed: %#v %v", run, replay, replayErr)
		}
	}

	full, err := second.Execute(derivedEnvelope(root, RuleDependsOn, LayerDeterministic, 1, 2))
	if err != nil {
		t.Fatal(err)
	}
	if len(full.Result.DerivedDeterministic) != 1 || full.Result.DerivedDeterministic[0].Object != target {
		t.Fatalf("larger work quota did not reach the canonical edge: %#v", full.Result)
	}
	permuted, err := first.Execute(derivedEnvelope(root, RuleDependsOn, LayerDeterministic, 1, 2))
	if err != nil || permuted.Hash != full.Hash {
		t.Fatalf("permuted quota replay changed: %#v %v", permuted, err)
	}
}

func TestEnvelopeTraversalQuotaStopsEdgeScanBeforeResultRows(t *testing.T) {
	graph := New()
	root := id("urn:query:quota:root")
	noise := id("urn:query:quota:noise")
	target := id("urn:query:quota:target")
	assertAdd(t, graph, NewFact(root, Used, noise))
	assertAdd(t, graph, NewFact(root, WasDerivedFrom, target))

	limited := traversalEnvelope(root, LayerDeterministic, 1, 1)
	limited.Relation = WasDerivedFrom
	response, err := graph.Execute(limited)
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Result.DeterministicPaths) != 0 {
		t.Fatalf("edge quota scanned beyond its limit: %#v", response.Result)
	}

	full := limited
	full.Limit = 2
	response, err = graph.Execute(full)
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Result.DeterministicPaths) != 1 || response.Result.DeterministicPaths[0].Last() != target {
		t.Fatalf("larger traversal quota did not reach the matching edge: %#v", response.Result)
	}
}
