package bodyexecution

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

const zeroInputSeed = `package constants
namespace constants
entity Integer id "urn:gooo:type:integer"
activity Seed() -> Integer id "urn:zero-input:seed" computes "return 7"
`

func TestZeroInputStageDoesNotInventACallerValue(t *testing.T) {
	source := []byte(zeroInputSeed + `activity Main(left Integer, right Integer) -> Integer id "urn:zero-input:main" computes "return left + right"
bind Seed.result -> Main.input0
`)
	suite := calledCompositionCases(t, `{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"Main.input1":5},"expected":{"Seed":7,"Main":12}}]}`)
	r, err := GenerateComposition(context.Background(), "constant-stage.gooo", source, suite, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Plan.Activities) != 2 || len(r.Plan.Activities[0].inputSlots()) != 0 {
		t.Fatal("constant stage acquired a phantom input", r.Plan)
	}
	inspection, err := InspectComposition(context.Background(), "constant-stage.gooo", source, "Main")
	if err != nil || len(inspection.CallerInputs) != 1 || inspection.CallerInputs[0].Key != "Main.input1" {
		t.Fatal("only the authored caller port is required", inspection, err)
	}
	native, err := ExecuteComposition(context.Background(), "constant-stage.gooo", source, r, suite, nativeTool())
	if err != nil || native.FinitePassed != 2 || native.FiniteTotal != 2 || native.ModelCalls != 0 {
		t.Fatal("source constant did not feed its declared target", native, err)
	}
	if len(native.Traces[0].Deliveries[0].Input) != 0 || len(native.Traces[0].Deliveries[0].Inputs) != 0 {
		t.Fatal("constant execution invented input provenance")
	}
	raw, _ := json.Marshal(r)
	if !strings.Contains(string(raw), `"zero_inputs":true`) {
		t.Fatal("zero declared arity was not recorded explicitly")
	}
	decoded, err := DecodeComposition(raw)
	if err != nil || VerifyComposition(context.Background(), "constant-stage.gooo", source, decoded) != nil {
		t.Fatal("zero-input declaration did not round trip", err)
	}
	decoded.Plan.Activities[0].ZeroInputs = false
	if VerifyComposition(context.Background(), "constant-stage.gooo", source, decoded) == nil {
		t.Fatal("removing the declared zero-input origin survived source replay")
	}
}

func TestZeroInputEntryRequiresAnExplicitEmptyInputObject(t *testing.T) {
	source := []byte(zeroInputSeed)
	inspection, err := InspectComposition(context.Background(), "constant.gooo", source, "Seed")
	if err != nil || inspection.CallerInputs == nil || len(inspection.CallerInputs) != 0 {
		t.Fatal("empty declared caller scope was lost", inspection, err)
	}
	template, err := CompositionInputTemplate(inspection)
	if err != nil || !strings.Contains(string(template), "{}") {
		t.Fatal("constant entry needs an empty object template", string(template), err)
	}
	for _, inputs := range []string{`{}`, `null`} {
		suite := calledCompositionCases(t, `{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":`+inputs+`,"expected":{"Seed":7}}]}`)
		r, err := GenerateComposition(context.Background(), "constant.gooo", source, suite, "")
		if inputs == "null" {
			if err == nil {
				t.Fatal("missing input context became a known empty argument tuple")
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		native, err := ExecuteComposition(context.Background(), "constant.gooo", source, r, suite, nativeTool())
		if err != nil || native.FinitePassed != 1 || native.ModelCalls != 0 {
			t.Fatal("constant entry did not execute its one oracle", native, err)
		}
	}
	inspection.CallerInputs = nil
	if _, err := CompositionInputTemplate(inspection); err == nil {
		t.Fatal("unknown caller scope became an empty argument template")
	}
	inspection.CallerInputs = []CompositionCallerInput{}
	inspection.Plan.Activities = nil
	if _, err := CompositionInputTemplate(inspection); err == nil {
		t.Fatal("an absent source plan became a known empty caller scope")
	}
}

func TestZeroInputFaultHasNoInventedReadAndRetainsIndependentOutput(t *testing.T) {
	source := []byte(strings.Replace(zeroInputSeed, "return 7", "let divisor = 0; return 7 / divisor", 1) +
		`activity Other() -> Integer id "urn:zero-input:other" computes "return 19"
activity Dependent(Integer) -> Integer id "urn:zero-input:dependent" computes "return input + 1"
bind Seed.result -> Dependent.input
`)
	suite := calledCompositionCases(t, `{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{},"expected":{"Seed":0,"Other":19,"Dependent":1}}]}`)
	r, err := GenerateComposition(context.Background(), "zero-fault.gooo", source, suite, "")
	if err != nil {
		t.Fatal(err)
	}
	native, err := ExecuteComposition(context.Background(), "zero-fault.gooo", source, r, suite, nativeTool())
	if err != nil {
		t.Fatal(err)
	}
	assertFaultJSON(t, native, 1, 1, 1)
	for i, delivery := range native.Traces[0].Deliveries {
		if r.Plan.Activities[i].ZeroInputs && (len(delivery.Input) != 0 || len(delivery.Inputs) != 0) {
			t.Fatal("constant fault or independent output acquired a phantom input")
		}
	}
}
