package bodyexecution

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestJointInputsRetainActualHelperArguments(t *testing.T) {
	source := []byte(strings.Replace(jointAmbiguousSource, "Choose(input)", "Choose(input * input)", 1))
	construction := jointCases(t, `[{"inputs":{"Main":3},"expected":{"Main":18}}]`)
	evaluation := jointCases(t, `[
 {"inputs":{"Main":-3},"expected":{"Main":18}},
 {"inputs":{"Main":4},"expected":{"Main":32}},
 {"inputs":{"Main":4},"expected":{"Main":32}},
 {"inputs":{"Main":0},"expected":{"Main":0}}]`)
	ctx := context.Background()
	prior, got, err := ConstructAndEvaluateJoint(ctx, "joint-inputs.gooo", source, construction, evaluation,
		JointOptions{EntryActivity: "Main", ProgramBudget: 2, GoBinary: nativeTool()})
	if err != nil {
		t.Fatal(err)
	}
	if len(prior.Attempts) != 2 || prior.Attempts[0].Runtime.FinitePassed != 0 || got.Runtime.FinitePassed != 4 {
		t.Fatal("construction failure or evaluation outputs changed")
	}
	for _, attempt := range prior.Attempts {
		trace := attempt.Runtime.Traces[0]
		if !trace.CalledInputsObserved || len(trace.Calls) != 1 || string(trace.Calls[0].Inputs[0]) != "9" {
			t.Fatal("construction lost its actual helper argument", trace)
		}
	}
	check := func(r JointEvaluation) {
		t.Helper()
		metric := r.Runtime.InputSeparation
		if metric.Status != "PASS" || metric.UniqueInputs != 3 || metric.DuplicateRows != 1 ||
			metric.OverlappingInputs != 2 || metric.DisjointInputs != 1 || metric.DisjointCasesPassed != 1 || metric.UnknownInputs != 0 {
			t.Fatal("new root hid prior helper/local exposure", metric)
		}
	}
	check(got)
	raw, _ := json.Marshal(prior)
	saved, err := DecodeJointConstruction(raw)
	if err != nil {
		t.Fatal(err)
	}
	replayed, err := ReplayJointComposition(ctx, "joint-inputs.gooo", source, saved, evaluation, nativeTool())
	if err != nil || !replayed.ConstructionReplayed || replayed.NewModelCalls != 0 {
		t.Fatal("history replay failed", err)
	}
	check(replayed)
}
