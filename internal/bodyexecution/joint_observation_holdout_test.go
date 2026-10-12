package bodyexecution

import (
	"context"
	"strings"
	"testing"
)

const jointObservedHoldoutSource = `package observedheld
namespace observedheld
entity Integer id "observedheld://integer"
activity Fee(Integer) -> Integer computes "return input - 67" assembling {
 choice "direction" operand_order at "0" intent "Subtract input from sixty seven."
 case "67" -> "0"
 observe { input "79" oracle "Reference" max_candidates "2" max_rounds "1" resolve_unique "true" }
 holdout_case "89" -> "-22"
 attempts "2"
}
activity Reference(Integer) -> Integer computes "return 67 - input"
activity Main(Integer) -> Integer computes "return Fee(input + 5)"
`

func TestJointPathHoldoutAndSourceObservationKeepDistinctEvidence(t *testing.T) {
	ctx := context.Background()
	cases := jointCases(t, `[{"inputs":{"Main":74},"expected":{"Main":12}}]`)
	for _, expected := range []string{"-22", "22"} {
		source := []byte(strings.Replace(jointObservedHoldoutSource, `"89" -> "-22"`, `"89" -> "`+expected+`"`, 1))
		r, err := ConstructJointComposition(ctx, "observed-held.gooo", source, cases,
			JointOptions{EntryActivity: "Main", ProgramBudget: 2, GoBinary: nativeTool()})
		if err != nil {
			t.Fatal(err)
		}
		if r.Decision != "PARTIAL_FINITE" || len(r.Attempts) != 2 || r.SelectedAttempt != 0 || len(r.FinalPathHoldouts) != 1 {
			t.Fatal("observation, caller and holdout decisions mixed", r)
		}
		first, second := r.Attempts[0], r.Attempts[1]
		if first.LocalPassed != 2 || first.LocalTotal != 2 || first.Runtime.FinitePassed != 0 ||
			second.LocalPassed != 1 || second.LocalTotal != 2 || second.Runtime.FinitePassed != 1 {
			t.Fatal("acquired answer lost or held answer entered selection", first, second)
		}
		held := r.FinalPathHoldouts[0]
		wantPassed := 0
		if expected == "-22" {
			wantPassed = 1
		}
		if held.Evaluation.Passed != wantPassed || held.Evaluation.Total != 1 ||
			len(held.CallerExposure.OtherInputs) != 1 || held.CallerExposure.OtherInputs[0] != 89 {
			t.Fatal("separate held score or caller exposure lost", held)
		}
		evaluation := jointCases(t, `[{"inputs":{"Main":103},"expected":{"Main":-41}}]`)
		replay, err := ReplayJointComposition(ctx, "observed-held.gooo", source, r, evaluation, nativeTool())
		if err != nil || !replay.ConstructionReplayed || replay.NewModelCalls != 0 || replay.Runtime.FinitePassed != 1 {
			t.Fatal("combined source contract failed replay", replay, err)
		}
	}
}
