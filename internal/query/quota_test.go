package query

import (
	"reflect"
	"testing"
)

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

func TestEnvelopeTraversalCycleLimitReplaysCanonicalPrefix(t *testing.T) {
	root := id("urn:query:cycle-limit:000-root")
	branch := id("urn:query:cycle-limit:100-branch")
	loop := id("urn:query:cycle-limit:200-loop")
	leaf := id("urn:query:cycle-limit:300-leaf")
	facts := []Fact{
		NewFact(root, WasDerivedFrom, branch),
		NewFact(branch, WasDerivedFrom, loop),
		NewFact(loop, WasDerivedFrom, root),
		NewFact(root, WasDerivedFrom, leaf),
	}
	first, second := New(), New()
	for _, fact := range facts {
		assertAdd(t, first, fact)
	}
	for index := len(facts) - 1; index >= 0; index-- {
		assertAdd(t, second, facts[index])
	}
	beforeCanonical, beforeHash, beforeNodes := first.Canonical(), first.StableHash(), first.Nodes()
	limitedRequest := traversalEnvelope(root, LayerDeterministic, 8, 2)
	limitedRequest.Relation = WasDerivedFrom
	limited, err := first.Execute(limitedRequest)
	if err != nil || len(limited.Result.DeterministicPaths) != 2 {
		t.Fatalf("bounded cycle traversal = %#v, err=%v", limited.Result, err)
	}
	if limited.Metadata.GraphHash != beforeHash {
		t.Fatalf("bounded cycle response changed graph hash: got %q want %q", limited.Metadata.GraphHash, beforeHash)
	}
	fullRequest := limitedRequest
	fullRequest.Limit = 10
	full, err := first.Execute(fullRequest)
	if err != nil || len(full.Result.DeterministicPaths) < 2 {
		t.Fatalf("expanded cycle traversal = %#v, err=%v", full.Result, err)
	}
	if full.Metadata.GraphHash != beforeHash {
		t.Fatalf("expanded cycle response changed graph hash: got %q want %q", full.Metadata.GraphHash, beforeHash)
	}
	if !reflect.DeepEqual(limited.Result.DeterministicPaths, full.Result.DeterministicPaths[:2]) {
		t.Fatalf("bounded cycle traversal changed canonical prefix: %#v vs %#v",
			limited.Result.DeterministicPaths, full.Result.DeterministicPaths[:2])
	}
	for _, path := range limited.Result.DeterministicPaths {
		if path.Depth() > 8 || hasRepeatedID(path.IDs) {
			t.Fatalf("bounded cycle traversal escaped simple path: %#v", path)
		}
	}
	for run := 0; run < 3; run++ {
		replay, replayErr := second.Execute(limitedRequest)
		if replayErr != nil || replay.Hash != limited.Hash {
			t.Fatalf("bounded cycle replay %d changed: %#v, err=%v", run, replay, replayErr)
		}
	}
	if first.Canonical() != beforeCanonical || first.StableHash() != beforeHash ||
		!reflect.DeepEqual(first.Nodes(), beforeNodes) {
		t.Fatal("bounded cycle traversal mutated graph authority")
	}
}

func TestEnvelopeTraversalInvalidRelationFailsClosedWithoutMutation(t *testing.T) {
	graph := New()
	root := id("urn:query:invalid-traversal:root")
	target := id("urn:query:invalid-traversal:target")
	assertAdd(t, graph, NewFact(root, WasDerivedFrom, target))
	request := traversalEnvelope(root, LayerDeterministic, 2, 2)
	request.Relation = Relation("gooo:unknown")
	beforeHash := graph.StableHash()
	response, err := graph.Execute(request)
	if err == nil || response.Error == nil || response.Error.Code != "unsupported_relation" {
		t.Fatalf("invalid traversal relation was not rejected: %#v, err=%v", response, err)
	}
	if response.Metadata.GraphHash != beforeHash || graph.StableHash() != beforeHash {
		t.Fatalf("invalid traversal relation changed graph hash: response=%q graph=%q want=%q",
			response.Metadata.GraphHash, graph.StableHash(), beforeHash)
	}
	if response.Hash == "" || response.Hash != response.CanonicalDigestValue() {
		t.Fatalf("invalid traversal rejection was not sealed: %#v", response)
	}
}
