package bodyexecution

import (
	"context"
	"encoding/json"
	"testing"
)

func TestJointSourceObservePreservesLocalAnswersAgainstCallerFeedback(t *testing.T) {
	source := []byte(`package retainedanswers
namespace retainedanswers
entity Integer id "retainedanswers://integer"
activity Fee(Integer) -> Integer computes "return input - 31" assembling {
 choice "direction" operand_order at "0" intent "Subtract the input from thirty one."
 case "31" -> "0"
 observe { input "43" oracle "Reference" max_candidates "2" max_rounds "1" resolve_unique "true" }
 attempts "2"
}
activity Reference(Integer) -> Integer computes "return 31 - input"
activity Main(Integer) -> Integer computes "return Fee(input)"
`)
	ctx := context.Background()
	cases := jointCases(t, `[{"inputs":{"Main":43},"expected":{"Main":12}}]`)
	r, err := ConstructJointComposition(ctx, "answers.gooo", source, cases,
		JointOptions{EntryActivity: "Main", ProgramBudget: 2, GoBinary: nativeTool()})
	if err != nil {
		t.Fatal(err)
	}
	if r.Decision != "PARTIAL_FINITE" || len(r.Attempts) != 2 || r.SelectedAttempt != 0 ||
		r.Attempts[0].LocalPassed != 2 || r.Attempts[0].LocalTotal != 2 || r.Attempts[0].Runtime.FinitePassed != 0 ||
		r.Attempts[1].LocalPassed != 1 || r.Attempts[1].LocalTotal != 2 || r.Attempts[1].Runtime.FinitePassed != 1 {
		t.Fatal("caller feedback displaced the acquired local answer", r)
	}
	for _, attempt := range r.Attempts {
		p := attempt.PathCandidates[0]
		if p.Observation == nil || p.Observation.AddedCases != 1 || p.Cases[1].Expected != -12 {
			t.Fatal("candidate lost its observation identity", p)
		}
	}
	evaluation := jointCases(t, `[{"inputs":{"Main":59},"expected":{"Main":-28}}]`)
	replay, err := ReplayJointComposition(ctx, "answers.gooo", source, r, evaluation, nativeTool())
	if err != nil || !replay.ConstructionReplayed || replay.NewModelCalls != 0 || replay.Runtime.FinitePassed != 1 {
		t.Fatal("retained obligations failed saved replay", replay, err)
	}
	raw, _ := json.Marshal(r)
	saved, err := DecodeJointConstruction(raw)
	if err != nil {
		t.Fatal(err)
	}
	saved.Attempts[0].PathCandidates[0].Observation.EffectiveCasesSHA256 = "changed"
	if _, err := ReplayJointComposition(ctx, "answers.gooo", source, saved, evaluation, nativeTool()); err == nil {
		t.Fatal("changed acquired-suite identity replayed")
	}
}
