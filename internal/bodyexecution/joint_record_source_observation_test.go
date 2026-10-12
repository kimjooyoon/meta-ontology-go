package bodyexecution

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestJointRecordObservationKeepsLocalAnswersAgainstCallerFeedback(t *testing.T) {
	source := strings.Replace(jointRecordHoldoutSource, ` attempts "2"`,
		` observe { value_input "[17]" oracle "Expected" max_candidates "2" max_rounds "1" resolve_unique "true" }
 attempts "2"`, 1)
	source += `activity Expected(Integer) -> Box computes "return Box{value: 0 - input}"` + "\n"
	ctx := context.Background()
	cases := jointCases(t, `[{"inputs":{"Main":14},"expected":{"Main":17}}]`)
	r, err := ConstructJointComposition(ctx, "recordanswers.gooo", []byte(source), cases,
		JointOptions{EntryActivity: "Main", ProgramBudget: 2, GoBinary: nativeTool()})
	if err != nil {
		t.Fatal(err)
	}
	if r.Decision != "PARTIAL_FINITE" || r.SelectedAttempt != 0 || len(r.Attempts) != 2 ||
		r.Attempts[0].LocalPassed != 2 || r.Attempts[0].LocalTotal != 2 || r.Attempts[0].Runtime.FinitePassed != 0 ||
		r.Attempts[1].LocalPassed != 1 || r.Attempts[1].LocalTotal != 2 || r.Attempts[1].Runtime.FinitePassed != 1 {
		t.Fatal(r)
	}
	for _, a := range r.Attempts {
		c := a.Candidates[0]
		if c.Observation == nil || c.Observation.AddedCases != 1 || string(c.Cases[1].Expected) != `{"value":-17}` {
			t.Fatal(c)
		}
	}
	if r.FinalRecordHoldouts[0].Evaluation.Passed != 1 {
		t.Fatal("held result lost", r.FinalRecordHoldouts)
	}
	raw, _ := json.MarshalIndent(r, "", "  ")
	saved, err := DecodeJointConstruction(raw)
	if err != nil {
		t.Fatal(err)
	}
	evaluation := jointCases(t, `[{"inputs":{"Main":31},"expected":{"Main":-34}}]`)
	replayed, err := ReplayJointComposition(ctx, "recordanswers.gooo", []byte(source), saved, evaluation, nativeTool())
	if err != nil || !replayed.ConstructionReplayed || replayed.NewModelCalls != 0 || replayed.Runtime.FinitePassed != 1 {
		t.Fatal(err, replayed)
	}
	saved.Attempts[0].Candidates[0].Observation.EffectiveCasesSHA256 = "changed"
	if _, err = ReplayJointComposition(ctx, "recordanswers.gooo", []byte(source), saved, evaluation, nativeTool()); err == nil {
		t.Fatal("forged retained record answers replayed")
	}
}

func TestJointRecordObservationCoexistsWithIntegerQuestions(t *testing.T) {
	source := strings.Replace(jointRecordHoldoutSource, ` attempts "2"`,
		` observe { value_input "[17]" oracle "Expected" max_candidates "2" max_rounds "1" resolve_unique "true" }
 attempts "2"`, 1)
	source = strings.Replace(source, `return result.value`, `return Adjust(result.value)`, 1)
	source += `activity Expected(Integer) -> Box computes "return Box{value: 0 - input}"
activity Adjust(Integer) -> Integer computes "return input - 1" assembling {
 choice "direction" operand_order at "0" intent "Subtract one."
 case "1" -> "0"
 observe { input "7" oracle "ScalarExpected" max_candidates "2" max_rounds "1" resolve_unique "true" }
 holdout_case "13" -> "12"
 attempts "2"
}
activity ScalarExpected(Integer) -> Integer computes "return input - 1"
`
	ctx := context.Background()
	cases := jointCases(t, `[{"inputs":{"Main":14},"expected":{"Main":-18}}]`)
	r, err := ConstructJointComposition(ctx, "mixedquestions.gooo", []byte(source), cases,
		JointOptions{EntryActivity: "Main", ProgramBudget: 4, GoBinary: nativeTool()})
	if err != nil {
		t.Fatal(err)
	}
	if r.Decision != "COMPLETE_FINITE" || len(r.Attempts) != 1 || r.Attempts[0].LocalTotal != 4 ||
		len(r.FinalRecordHoldouts) != 1 || len(r.FinalPathHoldouts) != 1 {
		t.Fatal(r)
	}
	evaluation := jointCases(t, `[{"inputs":{"Main":31},"expected":{"Main":-35}}]`)
	replay, err := ReplayJointComposition(ctx, "mixedquestions.gooo", []byte(source), r, evaluation, nativeTool())
	if err != nil || !replay.ConstructionReplayed || replay.NewModelCalls != 0 || replay.Runtime.FinitePassed != 1 {
		t.Fatal(err, replay)
	}
}
