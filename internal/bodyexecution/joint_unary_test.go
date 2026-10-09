package bodyexecution

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestJointTypedUnaryCallerConstructionAndReplay(t *testing.T) {
	source := []byte(strings.Replace(jointPathsSource, "return 0 - input", "return -input", 1))
	ctx := context.Background()
	cases := jointCases(t, `[{"inputs":{"Main":3},"expected":{"Main":6}}]`)
	prior, err := ConstructJointComposition(ctx, "unary.gooo", source, cases,
		JointOptions{EntryActivity: "Main", ProgramBudget: 2, GoBinary: nativeTool()})
	if err != nil {
		t.Fatal(err)
	}
	if len(prior.Attempts) != 2 || prior.Decision != "COMPLETE_FINITE" || prior.Attempts[0].Runtime.FinitePassed != 0 {
		t.Fatal("unary branch was not reconsidered", prior)
	}
	raw, err := json.Marshal(prior)
	if err != nil {
		t.Fatal(err)
	}
	saved, err := DecodeJointConstruction(raw)
	if err != nil {
		t.Fatal(err)
	}
	evaluation := jointCases(t, `[
		{"inputs":{"Main":9007199254740993},"expected":{"Main":18014398509481986}},
		{"inputs":{"Main":-5},"expected":{"Main":0}},
		{"inputs":{"Main":-9223372036854775808},"expected":{"Main":0}},
		{"inputs":{"Main":9223372036854775807},"expected":{"Main":-2}}]`)
	replay, err := ReplayJointComposition(ctx, "unary.gooo", source, saved, evaluation, nativeTool())
	if err != nil || replay.Runtime.FinitePassed != 4 || replay.NewModelCalls != 0 || !replay.ConstructionReplayed {
		t.Fatal("unary saved replay or int64 boundary changed", replay, err)
	}
}
