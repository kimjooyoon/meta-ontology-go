package bodycodegen

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
)

func TestSourceFillCandidateRetainsTrainingAndHoldout(t *testing.T) {
	source, err := os.ReadFile("../../examples/caller-source-fill/budget.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	set, err := PlanSourceFillCandidates(ctx, "budget.gooo", source, "PlanBudget")
	if err != nil || len(set.Candidates) != 3 {
		t.Fatal(set, err)
	}
	for i, id := range []string{"late_unbounded", "early_wrong_cap", "bounded"} {
		r, selected, err := RealizeSourceFillCandidate(ctx, "budget.gooo", source, "PlanBudget", id)
		if err != nil {
			t.Fatal(err)
		}
		wantHoldout := 0
		if i == 2 {
			wantHoldout = 1
		}
		if r.Schema != "gooo/fill-candidate/v1" || r.CandidateID != id || r.CandidateCount != 3 ||
			r.PlanSHA256 != set.PlanSHA256 || r.InputSourceSHA256 != digest(source) || r.SelectedSourceSHA256 != digest(selected) ||
			r.TestCasesPassed != 1 || r.TestCasesTotal != 1 || r.HoldoutCasesPassed != wantHoldout || r.HoldoutCasesTotal != 1 ||
			len(r.HoleFills) != 2 || len(r.ValueCaseResults) != 1 || len(r.ValueHoldoutResults) != 1 {
			t.Fatal(r)
		}
		if strings.Contains(string(selected), "__GOOO_BODY_HOLE_") || strings.Contains(string(selected), "assembling") {
			t.Fatal("unresolved assignment")
		}
	}
	if _, _, err = RealizeSourceFillCandidate(ctx, "budget.gooo", source, "PlanBudget", "invented"); err == nil {
		t.Fatal("undeclared candidate accepted")
	}
	if _, err = PlanSourceFillCandidates(nil, "budget.gooo", source, "PlanBudget"); err == nil {
		t.Fatal("nil context accepted")
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, _, err = RealizeSourceFillCandidate(cancelled, "budget.gooo", source, "PlanBudget", "bounded"); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestSourceFillCandidateIntegerGroupingAndDerivedRecord(t *testing.T) {
	// Parentheses must survive replacement; a local mismatch remains observable.
	source := []byte("package f\nnamespace f\nentity Integer id \"f://integer\"\n" +
		`activity Pick(Integer) -> Integer computes "return 2 * __GOOO_BODY_HOLE_left__ + __GOOO_BODY_HOLE_right__" assembling {
 source_fill intent "Shift and scale." { hole "left" hole "right"
 candidate "shift" { fill "left" "input + 1" fill "right" "0" }
 candidate "identity" { fill "left" "input" fill "right" "0" } }
 case "1" -> "4"
 holdout_case "9007199254740993" -> "18014398509481988"
}`)
	r, _, err := RealizeSourceFillCandidate(context.Background(), "fill.gooo", source, "Pick", "shift")
	if err != nil || r.TestCasesPassed != 1 || r.HoldoutCasesPassed != 1 || r.HoldoutCaseResults[0].Actual != 18014398509481988 {
		t.Fatal(r, err)
	}
	r, _, err = RealizeSourceFillCandidate(context.Background(), "fill.gooo", source, "Pick", "identity")
	if err != nil || r.TestCasesPassed != 0 || r.TestCasesTotal != 1 || r.CaseResults[0].Expected != 4 {
		t.Fatal(r, err)
	}
	source, err = os.ReadFile("../../examples/body-codegen/source-ir-fill-record-tiny.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	set, err := PlanSourceFillCandidates(context.Background(), "derived.gooo", source, "ReviewCandidate")
	if err != nil || set.Generation == nil || len(set.Candidates) != 2 {
		t.Fatal(set, err)
	}
	for _, candidate := range set.Candidates {
		r, _, err := RealizeSourceFillCandidate(context.Background(), "derived.gooo", source, "ReviewCandidate", candidate.ID)
		if err != nil || r.Generation == nil || r.PlanSHA256 != set.PlanSHA256 || r.TestCasesTotal != 2 || r.HoldoutCasesTotal != 2 {
			t.Fatal(r, err)
		}
	}
}
