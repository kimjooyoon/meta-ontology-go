package bodyexecution

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

const jointRecordHoldoutSource = `package recordcaller
namespace recordcaller
entity Integer id "recordcaller://integer"
entity Box id "recordcaller://box" fields { field value id "recordcaller://box/value" type integer required one }
activity Choose(Integer) -> Box id "recordcaller://choose" computes "return Box{value: input}" assembling {
 choice "direction" field_value at "0" alternative "0 - input" intent "Negate the input."
 value_case "[0]" -> "{\"value\":0}"
 holdout_value_case "[5]" -> "{\"value\":-5}"
 attempts "2"
}
activity Main(Integer) -> Integer id "recordcaller://main" computes "let result = Choose(input + 3); return result.value"
`

func TestJointRecordHoldoutFinalBodyAndCallerExposure(t *testing.T) {
	ctx := context.Background()
	for _, input := range []string{"17", "2"} {
		expected := "-20"
		if input == "2" {
			expected = "-5"
		}
		source := []byte(jointRecordHoldoutSource)
		construction := jointCases(t, `[{"inputs":{"Main":`+input+`},"expected":{"Main":`+expected+`}}]`)
		r, err := ConstructJointComposition(ctx, "heldrecord.gooo", source, construction, JointOptions{EntryActivity: "Main", ProgramBudget: 2, GoBinary: nativeTool()})
		if err != nil {
			t.Fatal(err)
		}
		if r.Schema != jointRecordHoldoutSchema || r.SelectedAttempt != 1 || r.Decision != "COMPLETE_FINITE" || len(r.FinalRecordHoldouts) != 1 {
			t.Fatal(r)
		}
		held := r.FinalRecordHoldouts[0]
		if r.Initial.Preparations[0].Generation.Report.RecordAssembly.Holdout.Passed != 0 || held.Evaluation.Passed != 1 || held.Evaluation.Total != 1 {
			t.Fatal(held)
		}
		for _, a := range r.Attempts {
			if a.LocalPassed != 1 || a.LocalTotal != 1 {
				t.Fatal("held values entered caller candidate score", a)
			}
		}
		if input == "17" && (len(held.CallerExposure.OtherInputs) != 1 || len(held.CallerExposure.ObservedInputs) != 0) {
			t.Fatal(held.CallerExposure)
		}
		if input == "2" && (len(held.CallerExposure.ObservedInputs) != 1 || held.CallerExposure.ObservedInputs[0] != "[5]" || len(held.CallerExposure.OtherInputs) != 0) {
			t.Fatal(held.CallerExposure)
		}
		raw, _ := json.MarshalIndent(r, "", "  ")
		saved, err := DecodeJointConstruction(raw)
		if err != nil {
			t.Fatal(err)
		}
		evaluation := jointCases(t, `[{"inputs":{"Main":31},"expected":{"Main":-34}}]`)
		replay, err := ReplayJointComposition(ctx, "heldrecord.gooo", source, saved, evaluation, nativeTool())
		if err != nil || !replay.ConstructionReplayed || replay.NewModelCalls != 0 || replay.Runtime.FinitePassed != 1 {
			t.Fatal(err, replay)
		}
		saved.FinalRecordHoldouts = nil
		saved.Schema = jointSchema
		if _, err = ReplayJointComposition(ctx, "heldrecord.gooo", source, saved, evaluation, nativeTool()); err == nil {
			t.Fatal("omitted final record evaluation replayed")
		}
	}
}

func TestJointRecordHoldoutAnswersDoNotChangeCallerSelection(t *testing.T) {
	ctx := context.Background()
	construction := jointCases(t, `[{"inputs":{"Main":11},"expected":{"Main":-14}}]`)
	var mask int = -1
	for _, expected := range []string{"-5", "5"} {
		source := []byte(strings.Replace(jointRecordHoldoutSource, `{\"value\":-5}`, `{\"value\":`+expected+`}`, 1))
		r, err := ConstructJointComposition(ctx, "answer.gooo", source, construction, JointOptions{EntryActivity: "Main", ProgramBudget: 2, GoBinary: nativeTool()})
		if err != nil {
			t.Fatal(err)
		}
		if mask < 0 {
			mask = r.SelectedAttempt
		}
		if r.SelectedAttempt != mask || len(r.Attempts) != 2 || r.Decision != "COMPLETE_FINITE" {
			t.Fatal("held answer changed caller selection")
		}
		want := 1
		if expected == "5" {
			want = 0
		}
		if r.FinalRecordHoldouts[0].Evaluation.Passed != want {
			t.Fatal(r.FinalRecordHoldouts)
		}
	}
}

func TestJointRecordAndPathHoldoutsCoexist(t *testing.T) {
	source := strings.Replace(jointRecordHoldoutSource,
		"return result.value\"", "return Adjust(result.value)\"", 1)
	source += `activity Adjust(Integer) -> Integer id "recordcaller://adjust" computes "return input - 1" assembling {
 choice "direction" operand_order at "0" intent "Subtract one."
 case "1" -> "0"
 holdout_case "7" -> "6"
 attempts "2"
}
`
	ctx := context.Background()
	cases := jointCases(t, `[{"inputs":{"Main":11},"expected":{"Main":-15}}]`)
	r, err := ConstructJointComposition(ctx, "mixedheld.gooo", []byte(source), cases,
		JointOptions{EntryActivity: "Main", ProgramBudget: 4, GoBinary: nativeTool()})
	if err != nil {
		t.Fatal(err)
	}
	if r.Schema != jointRecordHoldoutSchema || r.Decision != "COMPLETE_FINITE" || len(r.FinalRecordHoldouts) != 1 || len(r.FinalPathHoldouts) != 1 {
		t.Fatal("mixed construction lost a held profile", r)
	}
	if r.FinalRecordHoldouts[0].Evaluation.Passed != 1 || r.FinalPathHoldouts[0].Evaluation.Passed != 1 {
		t.Fatal("mixed final bodies were not evaluated independently")
	}
	evaluation := jointCases(t, `[{"inputs":{"Main":19},"expected":{"Main":-23}}]`)
	replayed, err := ReplayJointComposition(ctx, "mixedheld.gooo", []byte(source), r, evaluation, nativeTool())
	if err != nil || replayed.NewModelCalls != 0 || replayed.Runtime.FinitePassed != 1 {
		t.Fatal(err, replayed)
	}
	r.FinalPathHoldouts = nil
	if _, err = ReplayJointComposition(ctx, "mixedheld.gooo", []byte(source), r, evaluation, nativeTool()); err == nil {
		t.Fatal("record schema concealed missing path evaluation")
	}
}
