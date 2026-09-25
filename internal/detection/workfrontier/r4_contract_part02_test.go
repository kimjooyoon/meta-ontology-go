package workfrontier

import (
	"reflect"
	"testing"
)

func TestR4PermutationAndRootOrderIndependence(t *testing.T) {
	base := r4FixtureInput(t, "two-node-cycle")
	baseResult := EvaluateR4(base)
	permuted := base
	permuted.Pressures = reversePressures(base.Pressures)
	permuted.States = reverseStates(base.States)
	permuted.Paths = reversePaths(base.Paths)
	permuted.Rules = reverseRules(base.Rules)
	permuted.RootObligationIDs = []string{"obligation/root"}
	permutedResult := EvaluateR4(permuted)
	if !reflect.DeepEqual(baseResult, permutedResult) {
		t.Fatalf("permutation changed result:\nbase=%#v\npermuted=%#v", baseResult, permutedResult)
	}
	baseGraph, err := AnalyzeR4Graph(base)
	if err != nil {
		t.Fatal(err)
	}
	permutedGraph, err := AnalyzeR4Graph(permuted)
	if err != nil {
		t.Fatal(err)
	}
	if baseGraph.GraphDigest != permutedGraph.GraphDigest || baseGraph.SCCDigest != permutedGraph.SCCDigest || baseGraph.CondensationDigest != permutedGraph.CondensationDigest {
		t.Fatalf("permutation changed graph digests: %#v %#v", baseGraph, permutedGraph)
	}
	multiRoot := r4FixtureInput(t, "acyclic")
	multiRoot.RootObligationIDs = []string{"obligation/root", "obligation/child"}
	multiRoot, err = BindR4Payloads(multiRoot)
	if err != nil {
		t.Fatal(err)
	}
	reorderedRoots := multiRoot
	reorderedRoots.RootObligationIDs = []string{"obligation/child", "obligation/root"}
	if !reflect.DeepEqual(EvaluateR4(multiRoot), EvaluateR4(reorderedRoots)) {
		t.Fatal("root order changed the normalized result")
	}
	if got := FairBaseline(r4FixtureInput(t, "acyclic")); !reflect.DeepEqual(got, []string{"path/root"}) {
		t.Fatalf("fair baseline = %v", got)
	}
}
