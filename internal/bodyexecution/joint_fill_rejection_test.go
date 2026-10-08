package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestJointFillRejectionConsumesBudgetAndReplays(t *testing.T) {
	source, err := os.ReadFile("../../examples/caller-fill-rejection/budget.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	_, cases, evaluation := jointFillFixture(t)
	for _, budget := range []int{1, 3, 5} {
		r, err := ConstructJointComposition(context.Background(), "budget.gooo", source, cases, JointOptions{EntryActivity: "Main", ProgramBudget: budget, GoBinary: nativeTool()})
		if err != nil {
			t.Fatal(err)
		}
		if r.Schema != jointFillRejectionSchema || r.CandidateSpace != "5" || len(r.Attempts) != budget ||
			len(r.Initial.Preparations[0].Generation.Report.BodyFill.RejectedCandidates) != 2 {
			t.Fatal(r)
		}
		for i := 1; i < min(budget, 3); i++ {
			a := r.Attempts[i]
			if a.Rejection == nil || a.Rejection.Stage != "LOCAL_SOURCE_FILL" || a.FillCandidates[0].Rejection == nil ||
				a.LocalTotal != 0 || a.Runtime.Stage != "" || a.Runtime.FiniteTotal != 0 {
				t.Fatal(a)
			}
		}
		if (budget == 5) != (r.Decision == "COMPLETE_FINITE") {
			t.Fatal(r.Decision)
		}
		replayed, err := ReplayJointComposition(context.Background(), "budget.gooo", source, r, evaluation, nativeTool())
		if err != nil || replayed.NewModelCalls != 0 {
			t.Fatal(err)
		}
		if budget == 5 && replayed.Runtime.FinitePassed != 4 {
			t.Fatal(replayed)
		}
		if budget != 5 {
			continue
		}
		raw, _ := json.Marshal(r)
		for name, change := range map[string]func(*JointConstruction){
			"reason": func(r *JointConstruction) { r.Attempts[1].FillCandidates[0].Rejection.Reason += " changed" },
			"score":  func(r *JointConstruction) { r.Attempts[1].FillCandidates[0].TestCasesTotal = 1 },
			"native": func(r *JointConstruction) { r.Attempts[1].Runtime.FiniteTotal = 1 },
			"schema": func(r *JointConstruction) { r.Schema = jointFillSchema },
			"initial": func(r *JointConstruction) {
				r.Initial.Preparations[0].Generation.Report.BodyFill.RejectedCandidates = nil
			},
		} {
			t.Run(name, func(t *testing.T) {
				saved, err := DecodeJointConstruction(raw)
				if err != nil {
					t.Fatal(err)
				}
				change(&saved)
				if _, err = ReplayJointComposition(context.Background(), "budget.gooo", source, saved, evaluation, nativeTool()); err == nil {
					t.Fatal("altered rejection accepted")
				}
			})
		}
	}
}

func TestJointAllFillCandidatesRejectedBeforeModel(t *testing.T) {
	source, err := os.ReadFile("../../examples/caller-fill-rejection/budget.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	_, cases, _ := jointFillFixture(t)
	source = []byte(strings.Replace(string(source), "let next = input.used + 1", "let next = missing + 1", 1))
	r, err := ConstructJointComposition(context.Background(), "budget.gooo", source, cases, JointOptions{EntryActivity: "Main", ProgramBudget: 5, FillModelPath: "missing.json"})
	if err == nil || r.Initial.Stage != "BODY_PREFLIGHT" || r.Initial.FillFailure == nil || len(r.Initial.FillFailure.Rejected) != 5 ||
		r.Initial.FillModel != nil || len(r.Attempts) != 0 {
		t.Fatal(r, err)
	}
}

func TestJointFillRejectionKeepsRecordAndSearchPrefix(t *testing.T) {
	source, err := os.ReadFile("../../examples/caller-fill-rejection/budget.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	source = []byte(strings.Split(string(source), "activity Main(")[0] + `
activity Choose(Integer) -> Integer computes "return __GOOO_BODY_HOLE_value__" assembling {
 search hole "value" grammar "integer-offset-constant/v1" intent "Keep an integer." max_candidates "4"
 case "0" -> "0" attempts "3"
}
entity Box id "budgetplan://box" fields { field value id "budgetplan://box/value" type integer required one }
activity Pick(Integer) -> Box computes "return Box{value:input}" assembling {
 choice "double" field_value at "0" alternative "input * 2" intent "Double the input."
 value_case "[0]" -> "{\"value\":0}" attempts "2"
}
activity Main(Work) -> Integer computes "let p = Pick(input.used); let n = Choose(p.value); let b = PlanBudget(Work{used:n,limit:input.limit}); return b.next"
`)
	cases := jointCases(t, `[{"inputs":{"Main":{"used":0,"limit":8}},"expected":{"Main":1}}]`)
	initial, err := GenerateCompositionWithOptions(context.Background(), "prefix.gooo", source, cases, CompositionOptions{EntryActivity: "Main"})
	if err != nil {
		t.Fatal(err)
	}
	slots, _, err := jointSlots(context.Background(), "prefix.gooo", source, initial)
	if err != nil {
		t.Fatal(err)
	}
	var record, search, fill jointSlot
	for _, s := range slots {
		if len(s.fillIDs) > 0 {
			fill = s
		} else if len(s.searchIDs) > 0 {
			search = s
		} else {
			record = s
		}
	}
	a, selected, _, err := materializeJoint(context.Background(), "prefix.gooo", source, cases, "Main", []jointSlot{record, search, fill}, []uint16{0, search.initial, 1})
	if err != nil || selected != nil || a.Rejection == nil || a.Rejection.Slot != 2 || len(a.Candidates) != 1 || len(a.SearchCandidates) != 1 ||
		len(a.FillCandidates) != 1 || a.LocalPassed != 2 || a.LocalTotal != 2 || a.FillCandidates[0].Rejection == nil || a.Runtime.Stage != "" {
		t.Fatal(a, err)
	}
}
