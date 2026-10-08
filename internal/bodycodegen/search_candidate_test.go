package bodycodegen

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestSourceSearchCandidateUsesSourceBoundedExpression(t *testing.T) {
	source, err := os.ReadFile("../../examples/caller-ir-search/main.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	set, err := PlanSourceSearchCandidates(ctx, "search.gooo", source, "Choose")
	if err != nil {
		t.Fatal(err)
	}
	if set.AttemptBudget != 5 || len(set.Candidates) != 5 || set.Generation.CandidatesOmitted != 0 {
		t.Fatal(set)
	}
	for i, expression := range []string{"input", "0"} {
		result, selected, err := RealizeSourceSearchCandidate(ctx, "search.gooo", source, "Choose", set.Candidates[i].ID)
		if err != nil {
			t.Fatal(err)
		}
		if result.Attempt.Expression != expression || result.Attempt.TestCasesPassed != 1 || result.Attempt.CaseResults[0].Actual != 0 || strings.Contains(string(selected), "assembling") {
			t.Fatal(result, string(selected))
		}
	}
	limited := []byte(strings.Replace(string(source), `attempts "5"`, `attempts "1"`, 1))
	if _, _, err := RealizeSourceSearchCandidate(ctx, "limited.gooo", limited, "Choose", set.Candidates[1].ID); err == nil {
		t.Fatal("source attempt limit bypassed")
	}
	if _, _, err := RealizeSourceSearchCandidate(ctx, "search.gooo", source, "Choose", "made_up"); err == nil {
		t.Fatal("unknown candidate accepted")
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := PlanSourceSearchCandidates(canceled, "search.gooo", source, "Choose"); err == nil {
		t.Fatal("cancellation ignored")
	}
	if _, err := PlanSourceSearchCandidates(nil, "search.gooo", source, "Choose"); err == nil {
		t.Fatal("nil context accepted")
	}
}
