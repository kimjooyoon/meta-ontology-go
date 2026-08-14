package query

import (
	"errors"
	"reflect"
	"sync"
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

func TestDatalogCandidateIsolationAndBudgetFailuresAreNotPasses(t *testing.T) {
	ir := workspaceIR(t, "billing", "billing://", true)
	graph, err := FromSemanticIR(ir)
	if err != nil {
		t.Fatal(err)
	}
	request := workspaceDatalogRequest()
	request.IncludeCandidates = true
	result, err := graph.EvaluateDatalog(request)
	if err != nil {
		t.Fatal(err)
	}
	for _, fact := range result.Derived {
		if fact.Object == id("billing://entity/external") || fact.Origin == DatalogCandidate {
			t.Fatalf("candidate entered rule closure: %#v", result)
		}
	}
	for _, row := range result.Rows {
		if row.Bindings["source"] == id("billing://entity/external") {
			t.Fatalf("candidate-only dependsOn row was returned: %#v", row)
		}
	}

	bounded := request
	bounded.MaxDepth = 1
	bounded.IncludeCandidates = false
	bounded.MaxWork = DefaultDatalogWork
	boundedResult, err := graph.EvaluateDatalog(bounded)
	if !errors.Is(err, ErrDatalogBudget) || boundedResult.Complete {
		t.Fatalf("depth budget result = %#v, err=%v", boundedResult, err)
	}
	workBounded := request
	workBounded.IncludeCandidates = false
	workBounded.MaxDepth = DefaultDatalogDepth
	workBounded.MaxWork = 1
	workResult, err := graph.EvaluateDatalog(workBounded)
	if !errors.Is(err, ErrDatalogBudget) || workResult.Complete {
		t.Fatalf("work budget result = %#v, err=%v", workResult, err)
	}
}

func TestDatalogReplayIsRaceSafe(t *testing.T) {
	ir := workspaceIR(t, "billing", "billing://", false)
	graph, err := FromSemanticIR(ir)
	if err != nil {
		t.Fatal(err)
	}
	request := workspaceDatalogRequest()
	want, err := graph.EvaluateDatalog(request)
	if err != nil {
		t.Fatal(err)
	}
	wantDigest, err := want.CanonicalDigest()
	if err != nil {
		t.Fatal(err)
	}

	const workers = 8
	const replays = 20
	var wait sync.WaitGroup
	errorsCh := make(chan error, workers*replays)
	for worker := 0; worker < workers; worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for replay := 0; replay < replays; replay++ {
				got, replayErr := graph.EvaluateDatalog(request)
				if replayErr != nil {
					errorsCh <- replayErr
					continue
				}
				digest, digestErr := got.CanonicalDigest()
				if digestErr != nil || digest != wantDigest {
					errorsCh <- errors.New("Datalog replay digest changed")
				}
			}
		}()
	}
	wait.Wait()
	close(errorsCh)
	for replayErr := range errorsCh {
		t.Fatal(replayErr)
	}
}

func workspaceDatalogRequest() DatalogQuery {
	return DatalogQuery{
		Patterns: []Atom{Triple("dependsOn", Variable("entity"), Variable("source"))},
		Rules: []Rule{
			{
				ID:   "02/transitive-depends-on/v1",
				Head: Triple("dependsOn", Variable("entity"), Variable("source")),
				Body: []Atom{
					Triple("wasDerivedFrom", Variable("entity"), Variable("middle")),
					Triple("dependsOn", Variable("middle"), Variable("source")),
				},
			},
			{
				ID:   "01/direct-depends-on/v1",
				Head: Triple("dependsOn", Variable("entity"), Variable("source")),
				Body: []Atom{Triple("wasDerivedFrom", Variable("entity"), Variable("source"))},
			},
			{
				ID:   "00/inverse-used-by/v1",
				Head: Triple("usedBy", Variable("entity"), Variable("activity")),
				Body: []Atom{Triple("used", Variable("activity"), Variable("entity"))},
			},
		},
		IncludeDerived: true,
		Limit:          10,
	}
}
