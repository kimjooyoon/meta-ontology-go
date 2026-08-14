package query

import (
	"errors"
	"reflect"
	"sync"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
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

	request := workspaceDatalogRequest()
	permuted := request
	permuted.Patterns = append([]Atom(nil), request.Patterns...)
	for left, right := 0, len(permuted.Patterns)-1; left < right; left, right = left+1, right-1 {
		permuted.Patterns[left], permuted.Patterns[right] = permuted.Patterns[right], permuted.Patterns[left]
	}
	permuted.Rules = append([]Rule(nil), request.Rules...)
	for left, right := 0, len(permuted.Rules)-1; left < right; left, right = left+1, right-1 {
		permuted.Rules[left], permuted.Rules[right] = permuted.Rules[right], permuted.Rules[left]
	}
	requestDigest, err := request.CanonicalDigest()
	if err != nil {
		t.Fatal(err)
	}
	permutedDigest, err := permuted.CanonicalDigest()
	if err != nil || requestDigest != permutedDigest {
		t.Fatalf("rule/pattern permutation changed request digest: %q/%q, err=%v", requestDigest, permutedDigest, err)
	}

	beforeCanonical, beforeHash := first.Canonical(), first.StableHash()
	usedBy := DatalogQuery{
		Patterns: []Atom{Triple("usedBy", Variable("entity"), Variable("activity"))},
		Rules:    []Rule{request.Rules[2]}, IncludeDerived: true, Limit: 10,
	}
	usedByResult, err := first.EvaluateDatalog(usedBy)
	if err != nil {
		t.Fatal(err)
	}
	if len(usedByResult.Rows) != 1 || len(usedByResult.Derived) != 1 ||
		usedByResult.Rows[0].Bindings["entity"] != id("billing://entity/order") ||
		usedByResult.Rows[0].Bindings["activity"] != id("billing://activity/pay") {
		t.Fatalf("usedBy result = %#v", usedByResult)
	}
	if usedByResult.Rows[0].Facts[0].Origin != DatalogDerived || usedByResult.Derived[0].Namespace != "billing" {
		t.Fatalf("usedBy authority metadata = %#v", usedByResult)
	}
	if first.Canonical() != beforeCanonical || first.StableHash() != beforeHash {
		t.Fatal("Datalog projection mutated the authority graph")
	}

	dependsOn, err := first.EvaluateDatalog(request)
	if err != nil {
		t.Fatal(err)
	}
	permutedResult, err := first.EvaluateDatalog(permuted)
	if err != nil {
		t.Fatal(err)
	}
	dependsDigest, err := dependsOn.CanonicalDigest()
	if err != nil {
		t.Fatal(err)
	}
	permutedResultDigest, err := permutedResult.CanonicalDigest()
	if err != nil || dependsDigest != permutedResultDigest {
		t.Fatalf("rule permutation changed result digest: %q/%q, err=%v", dependsDigest, permutedResultDigest, err)
	}
	if len(dependsOn.Derived) != 4 || len(dependsOn.Rows) != 3 || !dependsOn.Complete {
		t.Fatalf("dependsOn result = %#v", dependsOn)
	}
	transitiveFound := false
	for _, row := range dependsOn.Rows {
		if row.Bindings["entity"] == id("billing://entity/payment") &&
			row.Bindings["source"] == id("billing://entity/base") {
			transitiveFound = true
		}
	}
	if !transitiveFound {
		t.Fatalf("transitive dependsOn row missing = %#v", dependsOn.Rows)
	}

	secondResult, err := second.EvaluateDatalog(request)
	if err != nil {
		t.Fatal(err)
	}
	if len(secondResult.Rows) != 3 || secondResult.Rows[0].Facts[0].Namespace != "settlement" {
		t.Fatalf("settlement projection = %#v", secondResult)
	}
	if reflect.DeepEqual(dependsOn.Rows, secondResult.Rows) {
		t.Fatal("different namespaces produced identical identity rows")
	}
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

func workspaceIR(t *testing.T, namespace, prefix string, candidate bool) semantic.IR {
	t.Helper()
	ir := semantic.NewIR(namespace, semantic.Namespace(namespace))
	entities := []struct {
		id, name string
	}{
		{prefix + "entity/payment", "Payment"},
		{prefix + "entity/order", "Order"},
		{prefix + "entity/base", "Base"},
	}
	for _, entity := range entities {
		node, err := semantic.NewEntity(semantic.MustIdentity(entity.id), semantic.Namespace(namespace), entity.name)
		if err != nil {
			t.Fatal(err)
		}
		if err := ir.AddNode(node); err != nil {
			t.Fatal(err)
		}
	}
	activity, err := semantic.NewActivity(semantic.MustIdentity(prefix+"activity/pay"), semantic.Namespace(namespace), "Pay")
	if err != nil {
		t.Fatal(err)
	}
	if err := ir.AddNode(activity); err != nil {
		t.Fatal(err)
	}
	add := func(fact semantic.Fact) {
		t.Helper()
		if err := ir.AddFact(fact); err != nil {
			t.Fatal(err)
		}
	}
	add(semantic.NewUsedFact(activity.ID, semantic.MustIdentity(prefix+"entity/order")))
	add(semantic.NewWasDerivedFromFact(semantic.MustIdentity(prefix+"entity/payment"), semantic.MustIdentity(prefix+"entity/order")))
	add(semantic.NewWasDerivedFromFact(semantic.MustIdentity(prefix+"entity/order"), semantic.MustIdentity(prefix+"entity/base")))
	if candidate {
		addCandidate := semantic.NewCandidateFact(
			semantic.MustIdentity(prefix+"entity/payment"),
			semantic.WasDerivedFrom,
			semantic.MustIdentity(prefix+"entity/external"),
			"unresolved cross-context observation",
		)
		// Candidates must also have declared endpoints in a validated IR.
		external, err := semantic.NewEntity(semantic.MustIdentity(prefix+"entity/external"), semantic.Namespace(namespace), "External")
		if err != nil {
			t.Fatal(err)
		}
		if err := ir.AddNode(external); err != nil {
			t.Fatal(err)
		}
		if err := ir.AddCandidate(addCandidate); err != nil {
			t.Fatal(err)
		}
	}
	return ir
}
