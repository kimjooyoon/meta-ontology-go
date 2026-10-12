package bodyexecution

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestCriteriaChallengeCallerCannotEraseSharedCounterexample(t *testing.T) {
	source := strings.Replace(callerCriteriaSource, `alternative "0-input"`, `alternative "input+0"`, 1)
	source = strings.Replace(source, `attempts "2"`, `observe {
 criteria_input "[7]"
 criteria_strategy "distinguish_then_check"
 max_candidates "2"
 max_rounds "1"
 reuse_outputs "true"
 resolve_unique "true"
}
attempts "2"`, 1)
	construction := jointCases(t, `[{"inputs":{"Main":2},"expected":{"Main":5}}]`)
	evaluation := jointCases(t, `[{"inputs":{"Main":8},"expected":{"Main":11}}]`)
	r, evaluated, err := ConstructAndEvaluateJoint(context.Background(), "challenge.gooo", []byte(source), construction, evaluation,
		JointOptions{EntryActivity: "Main", ProgramBudget: 2, GoBinary: nativeTool()})
	if err != nil {
		t.Fatal(err)
	}
	if r.Decision != "PARTIAL_FINITE" || len(r.Attempts) != 2 || evaluated.Runtime.FinitePassed != 1 {
		t.Fatal(r)
	}
	for _, a := range r.Attempts {
		if a.Runtime.FinitePassed != 1 || a.CriteriaPassed != 1 || a.CriteriaFailed != 1 || a.CriteriaTotal != 2 {
			t.Fatal("caller success erased source counterexample", a)
		}
	}
	raw, _ := json.Marshal(r)
	saved, err := DecodeJointConstruction(raw)
	if err != nil {
		t.Fatal(err)
	}
	replayed, err := ReplayJointComposition(context.Background(), "challenge.gooo", []byte(source), saved, evaluation, nativeTool())
	if err != nil || !replayed.ConstructionReplayed || replayed.NewModelCalls != 0 {
		t.Fatal(err, replayed)
	}
}
