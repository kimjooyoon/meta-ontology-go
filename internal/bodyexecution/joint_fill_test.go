package bodyexecution

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
)

func jointFillFixture(t *testing.T) ([]byte, CompositionCases, CompositionCases) {
	t.Helper()
	read := func(name string) []byte {
		raw, err := os.ReadFile("../../examples/caller-source-fill/" + name)
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}
	construction, err := DecodeCompositionCases(read("construction-cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	evaluation, err := DecodeCompositionCases(read("evaluation-cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	return read("budget.gooo.fixture"), construction, evaluation
}

func TestJointFillCallerSelectsCompleteAssignment(t *testing.T) {
	source, cases, evaluation := jointFillFixture(t)
	prior, err := ConstructJointComposition(context.Background(), "budget.gooo", source, cases,
		JointOptions{EntryActivity: "Main", ProgramBudget: 3, GoBinary: nativeTool()})
	if err != nil {
		t.Fatal(err)
	}
	if prior.Schema != jointFillSchema || prior.CandidateSpace != "3" || len(prior.Attempts) != 3 || prior.SelectedAttempt != 2 ||
		prior.Decision != "COMPLETE_FINITE" || prior.CandidateKinds[0] != "source_fill_index" ||
		prior.Initial.Preparations[0].Generation.Report.BodyFill.SelectedCandidateID != "late_unbounded" {
		t.Fatal(prior)
	}
	for i, attempt := range prior.Attempts {
		if attempt.LocalPassed != 1 || attempt.LocalTotal != 1 || len(attempt.FillCandidates) != 1 || attempt.FillCandidates[0].HoldoutCasesTotal != 1 ||
			attempt.FillCandidates[0].HoldoutCasesPassed != i/2 {
			t.Fatal(attempt)
		}
	}
	raw, _ := json.Marshal(prior)
	saved, err := DecodeJointConstruction(raw)
	if err != nil {
		t.Fatal(err)
	}
	replay, err := ReplayJointComposition(context.Background(), "budget.gooo", source, saved, evaluation, nativeTool())
	if err != nil || replay.Runtime.FinitePassed != 4 || !replay.ConstructionReplayed || replay.NewModelCalls != 0 || replay.InputSeparation.OtherInputs != 4 {
		t.Fatal(replay, err)
	}
	changes := map[string]func(*JointConstruction){
		"candidate": func(r *JointConstruction) { r.Attempts[0].FillCandidates[0].CandidateID = "bounded" },
		"hole":      func(r *JointConstruction) { r.Attempts[0].FillCandidates[0].HoleFills[0].Expression = "999" },
		"plan":      func(r *JointConstruction) { r.Attempts[0].FillCandidates[0].PlanSHA256 = "invented" },
		"actual": func(r *JointConstruction) {
			r.Attempts[0].FillCandidates[0].ValueCaseResults[0].Actual = json.RawMessage(`{"next":0,"exhausted":false}`)
		},
		"holdout score": func(r *JointConstruction) { r.Attempts[0].FillCandidates[0].HoldoutCasesPassed = 1 },
		"holdout expected": func(r *JointConstruction) {
			r.Attempts[0].FillCandidates[0].ValueHoldoutResults[0].Expected = json.RawMessage(`{}`)
		},
		"missing":   func(r *JointConstruction) { r.Attempts[0].FillCandidates = nil },
		"kind":      func(r *JointConstruction) { r.CandidateKinds[0] = "source_search_index" },
		"version":   func(r *JointConstruction) { r.Schema = jointMixedSchema },
		"index":     func(r *JointConstruction) { r.Attempts[0].Masks[0] = 2 },
		"truncated": func(r *JointConstruction) { r.Attempts = r.Attempts[1:]; r.SelectedAttempt-- },
	}
	for name, change := range changes {
		t.Run(name, func(t *testing.T) {
			saved, err := DecodeJointConstruction(raw)
			if err != nil {
				t.Fatal(err)
			}
			change(&saved)
			if _, err = ReplayJointComposition(context.Background(), "budget.gooo", source, saved, evaluation, nativeTool()); err == nil {
				t.Fatal("changed fill history accepted")
			}
		})
	}
}

func TestJointFillBudgetConflictAndHoldoutIndependence(t *testing.T) {
	source, cases, _ := jointFillFixture(t)
	for _, tc := range []struct {
		name, from, to, decision                   string
		budget, attempts, selected, local, holdout int
	}{
		{"budget", "", "", "PARTIAL_FINITE", 2, 2, 0, 1, 0},
		{"holdout", `holdout_value_case "[{\"used\":9,\"limit\":8}]" -> "{\"next\":8,\"exhausted\":true}"`, `holdout_value_case "[{\"used\":9,\"limit\":8}]" -> "{\"next\":999,\"exhausted\":true}"`, "COMPLETE_FINITE", 3, 3, 2, 1, 0},
		{"local conflict", `value_case "[{\"used\":0,\"limit\":8}]" -> "{\"next\":1,\"exhausted\":false}"`, `value_case "[{\"used\":0,\"limit\":8}]" -> "{\"next\":999,\"exhausted\":false}"`, "PARTIAL_FINITE", 3, 3, 2, 0, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			changed := source
			if tc.from != "" {
				if !strings.Contains(string(source), tc.from) {
					t.Fatal("fixture replacement missing")
				}
				changed = []byte(strings.Replace(string(source), tc.from, tc.to, 1))
			}
			r, err := ConstructJointComposition(context.Background(), "budget.gooo", changed, cases, JointOptions{EntryActivity: "Main", ProgramBudget: tc.budget, GoBinary: nativeTool()})
			if err != nil {
				t.Fatal(err)
			}
			if r.Decision != tc.decision || len(r.Attempts) != tc.attempts || r.SelectedAttempt != tc.selected ||
				r.Attempts[r.SelectedAttempt].LocalPassed != tc.local || r.Attempts[r.SelectedAttempt].FillCandidates[0].HoldoutCasesPassed != tc.holdout {
				t.Fatal(r)
			}
			if _, err = ReplayJointComposition(context.Background(), "budget.gooo", changed, r, cases, nativeTool()); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestJointFillBindingAndPreflight(t *testing.T) {
	source, suite := sourceFillCompositionFixture(t)
	r, err := ConstructJointComposition(context.Background(), "bind.gooo", source, suite, JointOptions{ProgramBudget: 3, GoBinary: nativeTool()})
	if err != nil || r.Schema != jointFillSchema || len(r.Attempts) != 1 || len(r.Attempts[0].FillCandidates) != 2 || r.Attempts[0].LocalTotal != 6 {
		t.Fatal(r, err)
	}
	if _, err = ReplayJointComposition(context.Background(), "bind.gooo", source, r, suite, nativeTool()); err != nil {
		t.Fatal(err)
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = ConstructJointComposition(cancelled, "bind.gooo", source, suite, JointOptions{ProgramBudget: 3, FillModelPath: "missing.json"}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	bad := []byte(strings.Replace(string(source), "return input > 0", "return missing > 0", 1))
	for _, input := range [][]byte{bad, source} {
		_, err = ConstructJointComposition(context.Background(), "bind.gooo", input, suite, JointOptions{ProgramBudget: 3, FillModelPath: "missing.json"})
		if err == nil || (string(input) == string(bad)) == strings.Contains(err.Error(), "load retained body-fill model") {
			t.Fatal("fill model flag or preflight order differs", err)
		}
	}
}

func TestJointFillSourceDerivedInteger(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/source-ir-fill-derived.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	cases := jointCases(t, `[{"inputs":{"Lift":4},"expected":{"Lift":5}}]`)
	r, err := ConstructJointComposition(context.Background(), "derived.gooo", source, cases, JointOptions{EntryActivity: "Lift", ProgramBudget: 16, GoBinary: nativeTool()})
	if err != nil || r.Schema != jointFillSchema || r.Decision != "COMPLETE_FINITE" || r.Attempts[0].FillCandidates[0].Generation == nil {
		t.Fatal(r, err)
	}
	if _, err = ReplayJointComposition(context.Background(), "derived.gooo", source, r, cases, nativeTool()); err != nil {
		t.Fatal(err)
	}
}

func TestJointFillSearchAndRecordWithRejectedPrefix(t *testing.T) {
	source, _, _ := jointFillFixture(t)
	source = []byte(strings.Split(string(source), "activity Main(")[0] + `
activity Choose(Integer) -> Integer computes "if input == 0 { return 0 }; return input / __GOOO_BODY_HOLE_value__" assembling {
 search hole "value" grammar "integer-offset-constant/v1" intent "Choose a caller value." max_candidates "8"
 case "0" -> "0" attempts "3"
}
entity Box id "budgetplan://box" fields { field value id "budgetplan://box/value" type integer required one }
activity Pick(Integer) -> Box computes "return Box{value:input}" assembling {
 choice "double" field_value at "0" alternative "input * 2" intent "Double the input."
 value_case "[0]" -> "{\"value\":0}" attempts "2"
}
activity Main(Work) -> Integer computes "let b = PlanBudget(input); let n = Choose(input.used); let p = Pick(b.next); if b.exhausted && n == -1 { return -p.value }; return p.value"
`)
	cases := jointCases(t, `[{"inputs":{"Main":{"used":8,"limit":8}},"expected":{"Main":-16}}]`)
	r, err := ConstructJointComposition(context.Background(), "mixed.gooo", source, cases, JointOptions{EntryActivity: "Main", ProgramBudget: 18, GoBinary: nativeTool()})
	if err != nil || r.Schema != jointFillSchema || r.CandidateSpace != "18" || r.Decision != "COMPLETE_FINITE" {
		t.Fatal(r, err)
	}
	rejections := 0
	for _, a := range r.Attempts {
		if a.Rejection != nil {
			rejections++
		}
	}
	if rejections == 0 {
		t.Fatal("missing search rejection")
	}
	if _, err = ReplayJointComposition(context.Background(), "mixed.gooo", source, r, cases, nativeTool()); err != nil {
		t.Fatal(err)
	}
	slots, _, err := jointSlots(context.Background(), "mixed.gooo", source, r.Initial)
	if err != nil {
		t.Fatal(err)
	}
	var fill, search, record jointSlot
	for _, s := range slots {
		if len(s.fillIDs) > 0 {
			fill = s
		} else if len(s.searchIDs) > 0 {
			search = s
		} else {
			record = s
		}
	}
	a, selected, _, err := materializeJoint(context.Background(), "mixed.gooo", source, cases, "Main",
		[]jointSlot{fill, search, record}, []uint16{0, 1, 0})
	if err != nil || a.Rejection == nil || a.Rejection.Slot != 1 || len(a.FillCandidates) != 1 || len(a.Candidates) != 0 || a.LocalTotal != 1 || selected != nil {
		t.Fatal(a, err)
	}
}
