package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestAssemblyPolicyDecisionsMatchNativeGoooExecution(t *testing.T) {
	ctx := context.Background()
	source, err := os.ReadFile("../../examples/body-codegen/record-field-assembly.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	policy, err := os.ReadFile("../../examples/assembly-explainer/main.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("../../examples/body-codegen/record-field-assembly-cases.json")
	if err != nil {
		t.Fatal(err)
	}
	suite, err := DecodeCompositionCases(raw)
	if err != nil {
		t.Fatal(err)
	}
	options := CompositionOptions{RecordPolicy: &bodycodegen.RecordAssemblyPolicy{Source: string(policy), Activity: "Explain"}}
	result, err := GenerateCompositionWithOptions(ctx, "record.gooo", source, suite, options)
	if err != nil {
		t.Fatal(err)
	}
	control := result.Steps[0].Generation.Report.RecordAssembly.Control
	observations := CompositionCases{Schema: "gooo/body-composition-cases/v1"}
	for _, decision := range control.Decisions {
		input, _ := json.Marshal(decision.Input)
		expected, _ := json.Marshal(map[string]string{"state": decision.State,
			"next_operation": decision.Operation, "message": decision.Message})
		observations.Cases = append(observations.Cases, CompositionCase{
			Inputs: map[string]json.RawMessage{"Explain": input}, Expected: map[string]json.RawMessage{"Explain": expected}})
	}
	compiled, err := GenerateComposition(ctx, "policy.gooo", policy, observations, "")
	if err != nil {
		t.Fatal(err)
	}
	native, err := ExecuteComposition(ctx, "policy.gooo", policy, compiled, observations, nativeTool())
	if err != nil || native.FinitePassed != len(control.Decisions) || !native.RuntimeReplayed {
		t.Fatal("interpreted policy differs from its native projection", err, native)
	}
	// Replay consumes the embedded policy; no policy path or model is supplied.
	native, err = ExecuteComposition(ctx, "record.gooo", source, result, suite, nativeTool())
	if err != nil || native.FinitePassed != 14 || native.ModelCalls != 0 || !native.ProjectionReplayed {
		t.Fatal("policy-controlled program did not replay natively", err, native)
	}
	control.Decisions[0].Continue = false
	if _, err = ExecuteComposition(ctx, "record.gooo", source, result, suite, nativeTool()); err == nil {
		t.Fatal("changed policy trace reached native execution")
	}
}
