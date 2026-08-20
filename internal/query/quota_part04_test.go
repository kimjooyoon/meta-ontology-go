package query

import (
	"reflect"
	"slices"
	"testing"
)

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
	for _, fact := range slices.Backward(facts) {
		assertAdd(t, second, fact)
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
	for run := range 3 {
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
