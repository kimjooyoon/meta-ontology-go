package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/criteriadecision"
)

const scalarNativeDecisionSource = `package nativedecision
namespace nativedecision
entity Integer id "urn:nativedecision:integer"
entity Boolean id "urn:nativedecision:boolean"
activity Fee(Integer) -> Integer id "urn:nativedecision:fee" computes "if input < 0 { return 0 } else { if input > 60 { return input*2 } else { return 120 } }" assembling {
 choice "cap" branch_layout at "1" intent "Cap the nonnegative doubled fee at 120."
 choice "negative" branch_layout at "0" intent "Keep the negative input fee at zero."
 criteria {
  predicate "required" activity "RequiredFee"
  input "[-9]" input "[12]"
  holdout_input "[-15]" holdout_input "[60]" holdout_input "[61]"
 }
 attempts "2"
}
activity RequiredFee(Integer,Integer) -> Boolean id "urn:nativedecision:required" computes "if input0 < 0 { return input1 == 0 } else { if input0 > 60 { return input1 == 120 } else { return input1 == input0*2 } }"
activity Main(Integer) -> Integer id "urn:nativedecision:main" computes "return Fee(input+1)"
`

func scalarNativeSyntheticArtifact(t *testing.T) string {
	t.Helper()
	var parameters [criteriadecision.ParameterCount]float32
	parameters[criteriadecision.ParameterCount-criteriadecision.HiddenDim-1] = 3
	metadata, weights, err := criteriadecision.EncodeArtifact(&parameters)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	for name, raw := range map[string][]byte{"metadata.json": metadata, "weights.bin": weights} {
		if err := os.WriteFile(filepath.Join(dir, name), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	return filepath.Join(dir, "metadata.json")
}

func TestScalarSourceModelEmitsAndExecutesCalledFeeNatively(t *testing.T) {
	suite, err := DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"Main":-4},"expected":{"Main":0}},{"inputs":{"Main":16},"expected":{"Main":34}},{"inputs":{"Main":100},"expected":{"Main":120}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	source := []byte(scalarNativeDecisionSource)
	composition, err := GenerateCompositionWithOptions(context.Background(), "native-scalar.gooo", source, suite,
		CompositionOptions{EntryActivity: "Main", ModelPath: scalarNativeSyntheticArtifact(t)})
	if err != nil || len(composition.Preparations) != 1 {
		t.Fatal("native source construction failed", err, composition.Stage)
	}
	r := composition.Preparations[0].Generation.Report.RecordAssembly
	if r == nil || r.ScalarPrediction == nil || r.ModelCalls != 1 || r.SelectedMask != 1 ||
		r.Criteria.Passed != 2 || r.Criteria.Total != 2 || r.CriteriaHoldout.Passed != 3 || r.CriteriaHoldout.Total != 3 {
		t.Fatal("called activity lost scalar model or criteria", r)
	}
	native, err := ExecuteComposition(context.Background(), "native-scalar.gooo", source, composition, suite, nativeTool())
	if err != nil || native.FinitePassed != 3 || native.FiniteTotal != 3 || native.ModelCalls != 0 ||
		!native.RuntimeReplayed || len(native.Runs) != 2 {
		t.Fatal("constructed scalar source did not execute natively", err, native.Stage, native.Failure)
	}
	logScalarNativeModelObservation(t, composition, native)
}

func logScalarNativeModelObservation(t *testing.T, c Composition, n CompositionRuntime) {
	t.Helper()
	r := c.Preparations[0].Generation.Report.RecordAssembly
	raw, err := json.Marshal(map[string]any{"source_sha256": c.OriginalSourceSHA256,
		"activity_id": c.Preparations[0].Generation.Report.ActivityID, "selected_mask": r.SelectedMask,
		"construction_model_calls": r.ModelCalls, "selection_passed": r.Criteria.Passed, "selection_total": r.Criteria.Total,
		"held_passed": r.CriteriaHoldout.Passed, "held_total": r.CriteriaHoldout.Total,
		"native_passed": n.FinitePassed, "native_total": n.FiniteTotal, "native_model_calls": n.ModelCalls,
		"native_processes": len(n.Runs), "native_elapsed_ns": n.ElapsedNS, "go_version": n.GoVersion,
		"generated_sha256": n.GeneratedSHA256, "executable_sha256": n.ExecutableSHA256})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("SCALAR_NATIVE_MODEL_OBSERVATION", string(raw))
}
