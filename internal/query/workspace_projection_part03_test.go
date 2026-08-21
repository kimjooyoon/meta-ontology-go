package query

import (
	"errors"
	"testing"
)

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
