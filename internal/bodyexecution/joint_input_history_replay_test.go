package bodyexecution

import (
	"context"
	"encoding/json"
	"testing"
)

func TestJointInputHistoryRejectsChangedObservations(t *testing.T) {
	ctx := context.Background()
	source := []byte(jointAmbiguousSource)
	cases := jointCases(t, `[{"inputs":{"Main":3},"expected":{"Main":6}}]`)
	prior, err := ConstructJointComposition(ctx, "input-binding.gooo", source, cases,
		JointOptions{EntryActivity: "Main", ProgramBudget: 2, GoBinary: nativeTool()})
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(prior)
	changes := map[string]func(*JointConstruction){
		"contract":         func(r *JointConstruction) { r.InputObservation = "unknown" },
		"removed contract": func(r *JointConstruction) { r.InputObservation = "" },
		"context":          func(r *JointConstruction) { r.Attempts[0].Runtime.CallContextSHA256 = "changed" },
		"argument":         func(r *JointConstruction) { r.Attempts[0].Runtime.Traces[0].Calls[0].Inputs[0] = json.RawMessage(`4`) },
		"observation":      func(r *JointConstruction) { r.Attempts[0].Runtime.Traces[0].CalledInputsObserved = false },
		"projection":       func(r *JointConstruction) { r.Attempts[0].Runtime.ObservedProjectionSHA256 = "changed" },
	}
	for name, change := range changes {
		t.Run(name, func(t *testing.T) {
			altered, err := DecodeJointConstruction(raw)
			if err != nil {
				t.Fatal(err)
			}
			change(&altered)
			before := compositionDigest(altered)
			got, err := ReplayJointComposition(ctx, "input-binding.gooo", source, altered, cases, nativeTool())
			if err == nil || got.ConstructionReplayed || got.Runtime.Stage != "" || compositionDigest(altered) != before {
				t.Fatal("changed history reached evaluation or rewrote its source", err)
			}
		})
	}
}

func TestJointInputHistoryPreservesLegacyUnknown(t *testing.T) {
	ctx := context.Background()
	source := []byte(jointAmbiguousSource)
	cases := jointCases(t, `[{"inputs":{"Main":3},"expected":{"Main":6}}]`)
	prior, err := ConstructJointComposition(ctx, "legacy-inputs.gooo", source, cases,
		JointOptions{EntryActivity: "Main", ProgramBudget: 2, GoBinary: nativeTool()})
	if err != nil {
		t.Fatal(err)
	}
	// Independently execute this new synthetic fixture through the legacy plain
	// projection route. No historical experiment or saved public receipt is run.
	prior.InputObservation = ""
	slots, _, err := jointSlots(ctx, "legacy-inputs.gooo", source, prior.Initial)
	if err != nil {
		t.Fatal(err)
	}
	for i := range prior.Attempts {
		_, selected, program, err := materializeJoint(ctx, "legacy-inputs.gooo", source, cases, "Main", slots, prior.Attempts[i].Masks)
		if err != nil {
			t.Fatal(err)
		}
		prior.Attempts[i].Runtime, err = ExecuteComposition(ctx, "legacy-inputs.gooo", selected, program, cases, nativeTool())
		if err != nil {
			t.Fatal(err)
		}
		prior.Attempts[i].Runtime.InputSeparation.Scope = "unique root-input tuples; actual inputs at graph activities and constructed helper calls compared with source selection/holdout cases and recorded probes; all supplied expectations across duplicate rows must match; model-training exposure is unknown"
	}
	before := compositionDigest(prior)
	got, err := ReplayJointComposition(ctx, "legacy-inputs.gooo", source, prior, cases, nativeTool())
	if err != nil || !got.ConstructionReplayed || got.Runtime.InputSeparation.Status != "UNKNOWN" ||
		got.Runtime.CallContextSHA256 != "" || got.Runtime.ConstructionInputHistorySHA256 != "" || compositionDigest(prior) != before {
		t.Fatal("legacy history was rewritten or given invented exposure", got, err)
	}
}

func TestJointRuntimeInputComparisonUsesMeasurements(t *testing.T) {
	a := CompositionRuntime{InputSeparation: CompositionInputSeparation{Status: "UNKNOWN", Reason: "UNCLASSIFIED_INPUTS",
		UniqueInputs: 2, UnknownInputs: 2, Scope: "earlier description"}}
	b := a
	b.InputSeparation.Scope = "clarified description"
	if !sameJointRuntime(a, b) {
		t.Fatal("descriptive wording changed an identical observation")
	}
	for _, change := range []func(*CompositionRuntime){
		func(r *CompositionRuntime) { r.InputSeparation.UniqueInputs++ },
		func(r *CompositionRuntime) { r.InputSeparation.Status = "PASS" },
		func(r *CompositionRuntime) { r.InputSeparation.Reason = "different" },
		func(r *CompositionRuntime) { r.CallContextSHA256 = "different" },
		func(r *CompositionRuntime) { r.ConstructionInputHistorySHA256 = "different" },
	} {
		b = a
		change(&b)
		if sameJointRuntime(a, b) {
			t.Fatal("input measurement or its source binding was ignored")
		}
	}
}
