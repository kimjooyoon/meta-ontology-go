package bodyexecution

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestCriteriaQuestionsCallerKeepsAcquiredObligationOnConflict(t *testing.T) {
	source := strings.Replace(callerCriteriaSource, `attempts "2"`, `observe {
 criteria_input "[7]"
 criteria_input "[-5]"
 max_candidates "2"
 max_rounds "1"
 reuse_outputs "true"
 resolve_unique "true"
}
attempts "2"`, 1)
	construction := jointCases(t, `[{"inputs":{"Main":4},"expected":{"Main":7}}]`)
	evaluation := jointCases(t, `[{"inputs":{"Main":8},"expected":{"Main":-11}}]`)
	r, evaluated, err := ConstructAndEvaluateJoint(context.Background(), "questions-caller.gooo", []byte(source), construction, evaluation,
		JointOptions{EntryActivity: "Main", ProgramBudget: 2, GoBinary: nativeTool()})
	if err != nil {
		t.Fatal(err)
	}
	if r.Decision != "PARTIAL_FINITE" || len(r.Attempts) != 2 || r.Attempts[0].CriteriaPassed != 2 || r.Attempts[0].CriteriaTotal != 2 ||
		r.Attempts[0].Runtime.FinitePassed != 0 || r.Attempts[1].CriteriaPassed != 1 || r.Attempts[1].CriteriaTotal != 2 ||
		r.Attempts[1].Runtime.FinitePassed != 1 || evaluated.Runtime.FinitePassed != 1 {
		t.Fatal("caller output erased the acquired predicate obligation", r)
	}
	raw, _ := json.MarshalIndent(r, "", "  ")
	saved, err := DecodeJointConstruction(raw)
	if err != nil {
		t.Fatal(err)
	}
	replayed, err := ReplayJointComposition(context.Background(), "questions-caller.gooo", []byte(source), saved, evaluation, nativeTool())
	if err != nil || !replayed.ConstructionReplayed || replayed.Runtime.FinitePassed != 1 || replayed.NewModelCalls != 0 {
		t.Fatal("question/caller replay", err)
	}
}
