package query

import (
	"slices"
	"testing"
)

func TestEnvelopeTraversalCandidateQuotaUsesRemainingRows(t *testing.T) {
	root := id("urn:query:candidate-quota:root")
	deterministicTarget := id("urn:query:candidate-quota:a-deterministic")
	candidateTarget := id("urn:query:candidate-quota:b-candidate")
	facts := []Fact{
		NewFact(root, WasDerivedFrom, deterministicTarget),
		NewCandidateFact(root, WasDerivedFrom, candidateTarget, "unresolved"),
	}
	first, second := New(), New()
	for _, fact := range facts {
		assertAdd(t, first, fact)
	}
	for _, fact := range slices.Backward(facts) {
		assertAdd(t, second, fact)
	}
	beforeHash := first.StableHash()
	request := traversalEnvelope(root, LayerAll, 1, 2)
	limited, err := first.Execute(request)
	if err != nil {
		t.Fatal(err)
	}
	if len(limited.Result.DeterministicPaths) != 1 || len(limited.Result.CandidatePaths) != 0 {
		t.Fatalf("candidate traversal exceeded remaining work quota: %#v", limited.Result)
	}
	if limited.Metadata.GraphHash != beforeHash || first.StableHash() != beforeHash {
		t.Fatalf("candidate quota changed graph hash: response=%q graph=%q want=%q",
			limited.Metadata.GraphHash, first.StableHash(), beforeHash)
	}
	permuted, err := second.Execute(request)
	if err != nil || permuted.Hash != limited.Hash {
		t.Fatalf("candidate quota replay changed: %#v, err=%v", permuted, err)
	}

	request.Limit = 3
	expanded, err := first.Execute(request)
	if err != nil || len(expanded.Result.CandidatePaths) != 1 {
		t.Fatalf("larger candidate work quota did not reach target: %#v, err=%v",
			expanded.Result, err)
	}
	if expanded.Metadata.GraphHash != beforeHash || first.StableHash() != beforeHash {
		t.Fatalf("expanded candidate query changed graph hash: response=%q graph=%q want=%q",
			expanded.Metadata.GraphHash, first.StableHash(), beforeHash)
	}
}
