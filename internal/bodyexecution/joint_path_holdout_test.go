package bodyexecution

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

const jointPathHoldoutSource = `package heldcaller
namespace heldcaller
entity Integer id "heldcaller://integer"
activity Fee(Integer) -> Integer computes "return input - 19" assembling {
 choice "direction" operand_order at "0" intent "Choose the subtraction direction."
 case "19" -> "0"
 holdout_case "23" -> "-4"
 attempts "2"
}
activity Main(Integer) -> Integer computes "return Fee(input + 3)"
`

func TestJointPathHoldoutEvaluatesFinalBodyWithoutAffectingSelection(t *testing.T) {
	ctx := context.Background()
	construction := jointCases(t, `[{"inputs":{"Main":26},"expected":{"Main":-10}}]`)
	for _, expected := range []string{"-4", "4"} {
		source := []byte(strings.Replace(jointPathHoldoutSource, `"23" -> "-4"`, `"23" -> "`+expected+`"`, 1))
		r, err := ConstructJointComposition(ctx, "held.gooo", source, construction, JointOptions{EntryActivity: "Main", ProgramBudget: 2, GoBinary: nativeTool()})
		if err != nil {
			t.Fatal(err)
		}
		if r.Schema != jointPathHoldoutSchema || r.Decision != "COMPLETE_FINITE" || r.SelectedAttempt != 1 || len(r.Attempts) != 2 || len(r.FinalPathHoldouts) != 1 {
			t.Fatal(r)
		}
		initial := r.Initial.Preparations[0].Generation.Report.BodyPaths.Holdout
		final := r.FinalPathHoldouts[0]
		if initial == nil || final.Evaluation == nil || initial.Passed == final.Evaluation.Passed ||
			final.Evaluation.Total != 1 || len(final.CallerExposure.OtherInputs) != 1 || len(final.CallerExposure.ObservedInputs) != 0 {
			t.Fatal(initial, final)
		}
		for _, attempt := range r.Attempts {
			if attempt.LocalTotal != 1 || attempt.LocalPassed != 1 {
				t.Fatal("holdout entered candidate score", attempt)
			}
		}
		evaluation := jointCases(t, `[{"inputs":{"Main":37},"expected":{"Main":-21}}]`)
		replayed, err := ReplayJointComposition(ctx, "held.gooo", source, r, evaluation, nativeTool())
		if err != nil || !replayed.ConstructionReplayed || replayed.NewModelCalls != 0 || replayed.Runtime.FinitePassed != 1 {
			t.Fatal(replayed, err)
		}
		raw, _ := json.Marshal(r)
		saved, err := DecodeJointConstruction(raw)
		if err != nil {
			t.Fatal(err)
		}
		saved.FinalPathHoldouts = nil
		saved.Schema = jointPathSchema
		if _, err := ReplayJointComposition(ctx, "held.gooo", source, saved, evaluation, nativeTool()); err == nil {
			t.Fatal("removed final holdouts replayed")
		}
	}
}

func TestJointPathHoldoutMeasuresActualCallerArgumentOverlap(t *testing.T) {
	ctx := context.Background()
	source := []byte(jointPathHoldoutSource)
	construction := jointCases(t, `[{"inputs":{"Main":20},"expected":{"Main":-4}}]`)
	r, err := ConstructJointComposition(ctx, "overlap.gooo", source, construction, JointOptions{EntryActivity: "Main", ProgramBudget: 2, GoBinary: nativeTool()})
	if err != nil {
		t.Fatal(err)
	}
	e := r.FinalPathHoldouts[0].CallerExposure
	if len(e.ObservedInputs) != 1 || e.ObservedInputs[0] != 23 || len(e.OtherInputs)+len(e.UnknownInputs) != 0 {
		t.Fatal("root input hid actual helper exposure", e)
	}
	calls, err := prepareJointCallContext(ctx, "overlap.gooo", source, r)
	if err != nil {
		t.Fatal(err)
	}
	r.Attempts[0].Runtime.CallContextSHA256 = "unavailable"
	r.Attempts[1].Runtime.CallContextSHA256 = "unavailable"
	sets, known := jointHoldoutCallerInputs(calls, r)
	unknown := measurePathHoldoutCallerExposure(calls, &sets, known, r.FinalPathHoldouts[0].Evaluation)
	if len(unknown.UnknownInputs) != 1 || len(unknown.OtherInputs) != 0 {
		t.Fatal("unobserved calls became independent input", unknown)
	}
}

func TestJointPathHoldoutIncludesLosingCallerAttempts(t *testing.T) {
	source := strings.Replace(jointPathHoldoutSource,
		`activity Main(Integer) -> Integer computes "return Fee(input + 3)"`,
		`activity Offset(Integer) -> Integer computes "return input - 2" assembling {
 choice "offset_direction" operand_order at "0" intent "Choose the offset direction."
 case "2" -> "0"
 attempts "2"
}
activity Main(Integer) -> Integer computes "let p = Offset(input); return Fee(p)"`, 1)
	cases := jointCases(t, `[{"inputs":{"Main":25},"expected":{"Main":42}}]`)
	r, err := ConstructJointComposition(context.Background(), "losing-input.gooo", []byte(source), cases,
		JointOptions{EntryActivity: "Main", ProgramBudget: 4, GoBinary: nativeTool()})
	if err != nil {
		t.Fatal(err)
	}
	if r.Decision != "COMPLETE_FINITE" || len(r.Attempts) != 4 || r.SelectedAttempt != 3 || len(r.FinalPathHoldouts) != 1 {
		t.Fatal("two source path choices did not compose", r)
	}
	e := r.FinalPathHoldouts[0].CallerExposure
	if len(e.ObservedInputs) != 1 || e.ObservedInputs[0] != 23 {
		t.Fatal("losing candidate exposure was forgotten", e)
	}
	feeID := r.FinalPathHoldouts[0].Evaluation.ActivityID
	for _, trace := range r.Attempts[r.SelectedAttempt].Runtime.Traces {
		for _, call := range trace.Calls {
			if call.ActivityID == feeID && string(call.Inputs[0]) != "-23" {
				t.Fatal("fixture did not distinguish final and earlier inputs", call)
			}
		}
	}
}

func TestJointPathHoldoutObservesDirectEntryArguments(t *testing.T) {
	source := []byte(jointPathHoldoutSource)
	cases := jointCases(t, `[{"inputs":{"Fee":23},"expected":{"Fee":-4}}]`)
	r, err := ConstructJointComposition(context.Background(), "direct-held.gooo", source, cases,
		JointOptions{EntryActivity: "Fee", ProgramBudget: 2, GoBinary: nativeTool()})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Initial.Preparations) != 0 || len(r.FinalPathHoldouts) != 1 ||
		len(r.FinalPathHoldouts[0].CallerExposure.ObservedInputs) != 1 || r.FinalPathHoldouts[0].CallerExposure.ObservedInputs[0] != 23 {
		t.Fatal("direct entry holdout exposure was lost", r)
	}
}
