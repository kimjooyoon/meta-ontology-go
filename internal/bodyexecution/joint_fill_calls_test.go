package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestJointFillLocalInputsRetainRejectedHelperExposure(t *testing.T) {
	read := func(name string) []byte {
		t.Helper()
		raw, err := os.ReadFile("../../examples/caller-fill-history/" + name)
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}
	cases := func(name string) CompositionCases {
		t.Helper()
		suite, err := DecodeCompositionCases(read(name))
		if err != nil {
			t.Fatal(err)
		}
		return suite
	}
	source := read("main.gooo.fixture")
	construction, evaluation := cases("construction-cases.json"), cases("evaluation-cases.json")
	ctx := context.Background()
	prior, got, err := ConstructAndEvaluateJoint(ctx, "fill.gooo", source, construction, evaluation,
		JointOptions{EntryActivity: "Main", ProgramBudget: 4, GoBinary: nativeTool()})
	if err != nil || prior.Decision != "COMPLETE_FINITE" || got.Runtime.FinitePassed != 3 || got.NewModelCalls != 0 {
		t.Fatal("constructed fill helper did not reach native evaluation", got, err)
	}
	fill := prior.Initial.Preparations[1].Generation.Report.BodyFill
	if fill == nil || fill.LocalCalls == nil || fill.SelectedCandidateID != "works" || len(fill.RejectedCandidates) != 1 {
		t.Fatal("rejected local assignment history missing", fill)
	}
	failed := fill.LocalCalls.Evaluations[0]
	if failed.EvaluationCompleted || len(failed.Calls) != 2 || string(failed.Calls[0].Inputs) != "[2]" ||
		string(failed.Calls[1].Inputs) != "[0]" {
		t.Fatal("failed assignment lost actual helper inputs", failed)
	}
	m := got.Runtime.InputSeparation
	if m.Status != "UNKNOWN" || m.Reason != "CONSTRUCTION_CALL_INPUTS_NOT_OBSERVED" ||
		m.OverlappingInputs != 2 || m.UnknownInputs != 1 || m.DisjointInputs != 0 {
		t.Fatal("failed-only exposure disappeared or incomplete capture gained independence", m)
	}
	raw, _ := json.MarshalIndent(prior, "", "  ")
	saved, err := DecodeJointConstruction(raw)
	if err != nil {
		t.Fatal(err)
	}
	replay, err := ReplayJointComposition(ctx, "fill.gooo", source, saved, evaluation, nativeTool())
	if err != nil || !replay.ConstructionReplayed || replay.NewModelCalls != 0 || replay.Runtime.FinitePassed != 3 ||
		!reflect.DeepEqual(m, replay.Runtime.InputSeparation) {
		t.Fatal("saved fill history changed during replay", replay, err)
	}
	saved.Initial.Preparations[1].Generation.Report.BodyFill.LocalCalls.Evaluations[0].Calls[1].Inputs = json.RawMessage("[99]")
	if _, err := ReplayJointComposition(ctx, "fill.gooo", source, saved, evaluation, nativeTool()); err == nil {
		t.Fatal("changed initial helper history was accepted")
	}
}
