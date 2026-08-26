package query

import (
	"reflect"
	"testing"
)

func TestSemanticWorkspaceProjectionIsScopedAndReplayable(t *testing.T) {
	firstIR := workspaceIR(t, "billing", "billing://", false)
	secondIR := workspaceIR(t, "settlement", "settlement://", false)
	first, err := FromSemanticIR(firstIR)
	if err != nil {
		t.Fatal(err)
	}
	second, err := FromSemanticIR(secondIR)
	if err != nil {
		t.Fatal(err)
	}
	assertWorkspaceScopedSearch(t, first, second)
	assertWorkspaceDatalog(t, first, second)
}
func assertWorkspaceScopedSearch(t *testing.T, first, second *Graph) {
	t.Helper()
	firstPayment, ok := first.NodeByName("billing", "Payment")
	if !ok || firstPayment.ID != id("billing://entity/payment") {
		t.Fatalf("billing Payment lookup = %#v, %v", firstPayment, ok)
	}
	secondPayment, ok := second.NodeByName("settlement", "Payment")
	if !ok || secondPayment.ID != id("settlement://entity/payment") {
		t.Fatalf("settlement Payment lookup = %#v, %v", secondPayment, ok)
	}
	if _, ok := first.NodeByName("settlement", "Payment"); ok {
		t.Fatal("cross-namespace display lookup resolved")
	}
	if _, ok := first.NodeByName("", "Payment"); ok {
		t.Fatal("unqualified display lookup resolved")
	}
	if first.Metadata().Namespace != "billing" || second.Metadata().Namespace != "settlement" ||
		first.StableHash() == second.StableHash() {
		t.Fatalf("projection scope identity = %#v/%#v, hashes=%q/%q", first.Metadata(), second.Metadata(), first.StableHash(), second.StableHash())
	}
	if got := first.Search("billing", "Payment"); len(got) != 1 || got[0].ID != firstPayment.ID {
		t.Fatalf("scoped search = %#v", got)
	}
}
func assertWorkspaceDatalog(t *testing.T, first, second *Graph) {
	t.Helper()
	request := workspaceDatalogRequest()
	permuted := permutedWorkspaceRequest(request)
	requestDigest, err := request.CanonicalDigest()
	if err != nil {
		t.Fatal(err)
	}
	permutedDigest, err := permuted.CanonicalDigest()
	if err != nil || requestDigest != permutedDigest {
		t.Fatalf("rule/pattern permutation changed request digest: %q/%q, err=%v", requestDigest, permutedDigest, err)
	}

	assertWorkspaceUsedBy(t, first, request.Rules[2])
	dependsOn := assertWorkspaceDependsOn(t, first, request, permuted)
	secondResult, err := second.EvaluateDatalog(request)
	if err != nil || len(secondResult.Rows) != 3 || secondResult.Rows[0].Facts[0].Namespace != "settlement" {
		t.Fatalf("settlement projection = %#v, err=%v", secondResult, err)
	}
	if reflect.DeepEqual(dependsOn.Rows, secondResult.Rows) {
		t.Fatal("different namespaces produced identical identity rows")
	}
}
