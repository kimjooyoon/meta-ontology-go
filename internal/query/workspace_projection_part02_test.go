package query

import (
	"testing"
)

func permutedWorkspaceRequest(request DatalogQuery) DatalogQuery {
	permuted := request
	permuted.Patterns = append([]Atom(nil), request.Patterns...)
	for left, right := 0, len(permuted.Patterns)-1; left < right; left, right = left+1, right-1 {
		permuted.Patterns[left], permuted.Patterns[right] = permuted.Patterns[right], permuted.Patterns[left]
	}
	permuted.Rules = append([]Rule(nil), request.Rules...)
	for left, right := 0, len(permuted.Rules)-1; left < right; left, right = left+1, right-1 {
		permuted.Rules[left], permuted.Rules[right] = permuted.Rules[right], permuted.Rules[left]
	}
	return permuted
}
func assertWorkspaceUsedBy(t *testing.T, graph *Graph, rule Rule) {
	t.Helper()
	request := DatalogQuery{
		Patterns: []Atom{Triple("usedBy", Variable("entity"), Variable("activity"))},
		Rules:    []Rule{rule}, IncludeDerived: true, Limit: 10,
	}
	result, err := graph.EvaluateDatalog(request)
	if err != nil || len(result.Rows) != 1 || len(result.Derived) != 1 {
		t.Fatalf("usedBy result = %#v, err=%v", result, err)
	}
	if result.Rows[0].Bindings["entity"] != id("billing://entity/order") ||
		result.Rows[0].Bindings["activity"] != id("billing://activity/pay") ||
		result.Rows[0].Facts[0].Origin != DatalogDerived || result.Derived[0].Namespace != "billing" {
		t.Fatalf("usedBy authority metadata = %#v", result)
	}
}
func assertWorkspaceDependsOn(t *testing.T, graph *Graph, request, permuted DatalogQuery) DatalogResult {
	t.Helper()
	beforeCanonical, beforeHash := graph.Canonical(), graph.StableHash()
	result, err := graph.EvaluateDatalog(request)
	if err != nil {
		t.Fatal(err)
	}
	permutedResult, err := graph.EvaluateDatalog(permuted)
	if err != nil {
		t.Fatal(err)
	}
	dependsDigest, err := result.CanonicalDigest()
	if err != nil {
		t.Fatal(err)
	}
	permutedDigest, err := permutedResult.CanonicalDigest()
	if err != nil || dependsDigest != permutedDigest {
		t.Fatalf("rule permutation changed result digest: %q/%q, err=%v", dependsDigest, permutedDigest, err)
	}
	if len(result.Derived) != 4 || len(result.Rows) != 3 || !result.Complete {
		t.Fatalf("dependsOn result = %#v", result)
	}
	if graph.Canonical() != beforeCanonical || graph.StableHash() != beforeHash {
		t.Fatal("Datalog projection mutated the authority graph")
	}
	for _, row := range result.Rows {
		if row.Bindings["entity"] == id("billing://entity/payment") &&
			row.Bindings["source"] == id("billing://entity/base") {
			return result
		}
	}
	t.Fatalf("transitive dependsOn row missing = %#v", result.Rows)
	return DatalogResult{}
}
