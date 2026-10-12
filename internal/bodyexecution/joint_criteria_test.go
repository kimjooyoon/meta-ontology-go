package bodyexecution

import (
	"context"
	"encoding/json"
	"testing"
)

const callerCriteriaSource = `package criteriacaller
namespace criteriacaller
entity Integer id "criteriacaller://integer"
entity Boolean id "criteriacaller://boolean"
entity Box id "criteriacaller://box" fields { field value id "criteriacaller://value" type integer required one }
activity Choose(Integer) -> Box computes "return Box{value:input}" assembling {
 choice "direction" field_value at "0" alternative "0-input" intent "Negate."
 criteria { predicate "negated" activity "Negated" input "[0]" holdout_input "[5]" }
 attempts "2"
}
activity Negated(Integer,Box) -> Boolean computes "return input1.value == 0-input0"
activity Main(Integer) -> Integer computes "let result = Choose(input + 3); return result.value"`

func TestCriteriaCallerConstructionPreservesFinalHoldoutAndReplay(t *testing.T) {
	ctx := context.Background()
	source := []byte(callerCriteriaSource)
	construction := jointCases(t, `[{"inputs":{"Main":2},"expected":{"Main":-5}}]`)
	evaluation := jointCases(t, `[{"inputs":{"Main":31},"expected":{"Main":-34}}]`)
	r, evaluated, err := ConstructAndEvaluateJoint(ctx, "caller-criteria.gooo", source, construction, evaluation,
		JointOptions{EntryActivity: "Main", ProgramBudget: 2, GoBinary: nativeTool()})
	if err != nil {
		t.Fatal(err)
	}
	if r.Decision != "COMPLETE_FINITE" || r.Schema != jointCriteriaSchema || len(r.Attempts) != 2 || r.SelectedAttempt != 1 ||
		r.Attempts[1].LocalTotal != 0 || r.Attempts[1].CriteriaPassed != 1 || r.Attempts[1].CriteriaTotal != 1 ||
		len(r.FinalCriteriaHoldouts) != 1 || r.FinalCriteriaHoldouts[0].Evaluation.Passed != 1 ||
		evaluated.Runtime.FinitePassed != 1 || evaluated.NewModelCalls != 0 {
		t.Fatalf("caller criteria construction differs: %+v", r)
	}
	initial := r.Initial.Preparations[0].Generation.Report.RecordAssembly
	if initial.CriteriaHoldout.Failed != 1 || r.FinalCriteriaHoldouts[0].Evaluation.Inputs[0].ActualJSON != `{"value":-5}` ||
		len(r.FinalCriteriaHoldouts[0].CallerExposure.ObservedInputs) != 1 {
		t.Fatal("initial/final evidence or caller exposure lost", r.FinalCriteriaHoldouts)
	}
	raw, _ := json.MarshalIndent(r, "", "  ")
	saved, err := DecodeJointConstruction(raw)
	if err != nil {
		t.Fatal(err)
	}
	replayed, err := ReplayJointComposition(ctx, "caller-criteria.gooo", source, saved, evaluation, nativeTool())
	if err != nil || !replayed.ConstructionReplayed || replayed.Runtime.FinitePassed != 1 {
		t.Fatal("caller criteria replay", err)
	}
	// Local comparison runs before native execution and rejects missing counts.
	saved.Attempts[0].CriteriaTotal = 0
	if _, err := ReplayJointComposition(ctx, "caller-criteria.gooo", source, saved, evaluation, nativeTool()); err == nil {
		t.Fatal("missing local criteria counts replayed")
	}
}

func TestCriteriaCallerSelectionCannotIgnoreUnmetLocalPredicates(t *testing.T) {
	bad := JointAttempt{CriteriaPassed: 1, CriteriaFailed: 1, CriteriaTotal: 2, Runtime: CompositionRuntime{FinitePassed: 1, FiniteTotal: 1}}
	good := JointAttempt{CriteriaPassed: 2, CriteriaTotal: 2, Runtime: CompositionRuntime{FinitePassed: 1, FiniteTotal: 1}}
	if jointComplete(bad) || !jointComplete(good) || !betterJoint(good, bad) {
		t.Fatal("caller match discarded local predicates")
	}
}
