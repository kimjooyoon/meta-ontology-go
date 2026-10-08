package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func jointRejectionSource(t *testing.T) []byte {
	t.Helper()
	source, err := os.ReadFile("../../examples/caller-search-rejection/main.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	return source
}

func TestJointSearchContinuesAfterRejectedExpression(t *testing.T) {
	source := jointRejectionSource(t)
	feedback := jointCases(t, `[{"inputs":{"Main":3},"expected":{"Main":-1}}]`)
	for _, budget := range []int{1, 2, 3, 5} {
		r, err := ConstructJointComposition(context.Background(), "rejection.gooo", source, feedback,
			JointOptions{EntryActivity: "Main", ProgramBudget: budget, GoBinary: nativeTool()})
		if err != nil {
			t.Fatal(err)
		}
		if len(r.Attempts) != min(budget, 3) || r.CandidateSpace != "5" {
			t.Fatal(r)
		}
		if budget == 1 && r.Schema != jointMixedSchema || budget > 1 && r.Schema != jointRejectionSchema {
			t.Fatal("rejection schema differs", r.Schema)
		}
		if budget > 1 {
			a := r.Attempts[1]
			if a.Rejection == nil || a.Rejection.Slot != 0 || a.Rejection.Stage != "LOCAL_SOURCE_SEARCH" || a.Rejection.Activity != "Choose" ||
				a.Rejection.Reason == "" || a.SearchCandidates[0].Attempt.Expression != "0" || a.LocalPassed != 0 || a.LocalTotal != 0 ||
				a.Runtime.Stage != "" || a.Runtime.FiniteTotal != 0 || jointComplete(a) || jointLocalComplete(a) {
				t.Fatal("rejected expression was scored or executed", a)
			}
		}
		if budget < 3 {
			if r.Decision != "PARTIAL_FINITE" || r.StopReason != "PROGRAM_BUDGET_EXHAUSTED" || r.SelectedAttempt != 0 {
				t.Fatal(r)
			}
		} else if r.Decision != "COMPLETE_FINITE" || r.SelectedAttempt != 2 || r.Attempts[2].SearchCandidates[0].Attempt.Expression != "-input" {
			t.Fatal(r)
		}
		if _, err = ReplayJointComposition(context.Background(), "rejection.gooo", source, r, feedback, nativeTool()); err != nil {
			t.Fatal("failed to replay bounded rejection history", err)
		}
	}
}

func TestJointSearchRejectionReplayAndExactEvaluation(t *testing.T) {
	source := jointRejectionSource(t)
	feedback := jointCases(t, `[{"inputs":{"Main":3},"expected":{"Main":-1}}]`)
	prior, err := ConstructJointComposition(context.Background(), "rejection.gooo", source, feedback,
		JointOptions{EntryActivity: "Main", ProgramBudget: 5, GoBinary: nativeTool()})
	if err != nil {
		t.Fatal(err)
	}
	evaluation := jointCases(t, `[{"inputs":{"Main":7},"expected":{"Main":-1}},{"inputs":{"Main":-5},"expected":{"Main":-1}},{"inputs":{"Main":0},"expected":{"Main":0}},{"inputs":{"Main":9007199254740993},"expected":{"Main":-1}}]`)
	raw, _ := json.Marshal(prior)
	saved, err := DecodeJointConstruction(raw)
	if err != nil {
		t.Fatal(err)
	}
	replay, err := ReplayJointComposition(context.Background(), "rejection.gooo", source, saved, evaluation, nativeTool())
	if err != nil || replay.Runtime.FinitePassed != 4 || !replay.ConstructionReplayed || replay.NewModelCalls != 0 || replay.InputSeparation.OtherInputs != 4 {
		t.Fatal(replay, err)
	}
	changes := map[string]func(*JointConstruction){
		"missing":   func(r *JointConstruction) { r.Attempts[1].Rejection = nil },
		"reason":    func(r *JointConstruction) { r.Attempts[1].Rejection.Reason = "invented" },
		"slot":      func(r *JointConstruction) { r.Attempts[1].Rejection.Slot = 1 },
		"selector":  func(r *JointConstruction) { r.Attempts[1].Masks[0] = 2 },
		"candidate": func(r *JointConstruction) { r.Attempts[1].SearchCandidates[0].Attempt.Expression = "1" },
		"scored":    func(r *JointConstruction) { r.Attempts[1].SearchCandidates[0].Attempt.ScoringCompleted = true },
		"native":    func(r *JointConstruction) { r.Attempts[1].Runtime.FiniteTotal = 1 },
		"selection": func(r *JointConstruction) { r.SelectedAttempt = 1 },
		"version":   func(r *JointConstruction) { r.Schema = jointMixedSchema },
		"truncated": func(r *JointConstruction) {
			r.Attempts = append(r.Attempts[:1], r.Attempts[2:]...)
			r.SelectedAttempt = 1
		},
	}
	for name, change := range changes {
		t.Run(name, func(t *testing.T) {
			r, err := DecodeJointConstruction(raw)
			if err != nil {
				t.Fatal(err)
			}
			change(&r)
			if _, err = ReplayJointComposition(context.Background(), "rejection.gooo", source, r, evaluation, nativeTool()); err == nil {
				t.Fatal("changed rejection history accepted")
			}
		})
	}
}

func TestJointSearchToolErrorsStopAndArithmeticFaultsContinue(t *testing.T) {
	source := jointRejectionSource(t)
	feedback := jointCases(t, `[{"inputs":{"Main":3},"expected":{"Main":-1}}]`)
	if _, err := ConstructJointComposition(context.Background(), "rejection.gooo", source, feedback,
		JointOptions{EntryActivity: "Main", ProgramBudget: 5, GoBinary: "missing-go-tool"}); err == nil {
		t.Fatal("missing tool hidden as a candidate rejection")
	}
	// Keep the original variable-zero counterexample. It now has a source-bound
	// native fault outcome, separate from a rejected local expression.
	changed := strings.Replace(string(source), "if input == 0 { return 0 }; return input / __GOOO_BODY_HOLE_value__",
		"let denominator = __GOOO_BODY_HOLE_value__; if input == 0 { return 0 }; return input / denominator", 1)
	r, err := ConstructJointComposition(context.Background(), "native.gooo", []byte(changed), feedback,
		JointOptions{EntryActivity: "Main", ProgramBudget: 5, GoBinary: nativeTool()})
	if err != nil || r.Failure != "" || len(r.Attempts) != 3 || r.Attempts[1].Rejection != nil ||
		!hasCompositionFault(r.Attempts[1].Runtime) || r.Schema != jointFaultSchema || r.SelectedAttempt != 2 ||
		r.Decision != "COMPLETE_FINITE" || !r.Attempts[1].Runtime.RuntimeReplayed {
		t.Fatal("native fault lost or hidden as a local candidate rejection", r, err)
	}
}

func TestJointRejectionRetainsScoredPrefix(t *testing.T) {
	source, err := os.ReadFile("../../examples/caller-search-rejection/mixed-model.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	cases := jointCases(t, `[{"inputs":{"Main":3},"expected":{"Main":15}}]`)
	initial, err := GenerateCompositionWithOptions(ctx, "prefix.gooo", source, cases, CompositionOptions{EntryActivity: "Main"})
	if err != nil {
		t.Fatal(err)
	}
	slots, _, err := jointSlots(ctx, "prefix.gooo", source, initial)
	if err != nil || len(slots) != 2 || slots[0].activity != "Choose" || slots[1].activity != "Pick" {
		t.Fatal(slots, err)
	}
	// Independent helpers can be realized in this order. The record was already
	// scored when the following search expression was rejected.
	a, sourceOut, _, err := materializeJoint(ctx, "prefix.gooo", source, cases, "Main",
		[]jointSlot{slots[1], slots[0]}, []uint16{0, 1})
	if err != nil || a.Rejection == nil || a.Rejection.Slot != 1 || a.LocalPassed != 1 || a.LocalTotal != 1 ||
		len(a.Candidates) != 1 || len(a.SearchCandidates) != 1 || sourceOut != nil || jointLocalComplete(a) {
		t.Fatal(a, err)
	}
}
