package bodyexecution

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
)

func inputJoinFixture(t *testing.T) ([]byte, CompositionCases) {
	t.Helper()
	source, err := os.ReadFile("../../examples/body-codegen/native-input-joins.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("../../examples/body-codegen/native-input-joins-cases.json")
	if err != nil {
		t.Fatal(err)
	}
	suite, err := DecodeCompositionCases(raw)
	if err != nil {
		t.Fatal(err)
	}
	return source, suite
}

func TestCompositionNativeRepeatedMixedAndPartlyBoundInputs(t *testing.T) {
	source, suite := inputJoinFixture(t)
	prior, err := GenerateComposition(context.Background(), "joins.gooo", source, suite, "")
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(prior)
	prior, err = DecodeComposition(raw)
	if err != nil {
		t.Fatal(err)
	}
	run, err := ExecuteComposition(context.Background(), "joins.gooo", source, prior, suite, nativeTool())
	if err != nil || run.FinitePassed != 49 || run.FiniteTotal != 49 || !run.RuntimeReplayed || run.ModelCalls != 0 {
		t.Fatalf("native inputs: %v %+v", err, run)
	}
	ports := 0
	for _, trace := range run.Traces {
		for i, entry := range trace.Deliveries {
			node := prior.Plan.Activities[i]
			if len(node.Inputs) == 0 {
				continue
			}
			if len(entry.Inputs) != len(node.Inputs) || len(entry.Input) != 0 {
				t.Fatal("multi-input trace has the wrong arity")
			}
			for p, input := range entry.Inputs {
				ports++
				if input.Port != node.Inputs[p].Port || input.EntityID != node.Inputs[p].EntityID {
					t.Fatal("trace source port identity differs")
				}
				from := node.Inputs[p].From
				if from >= 0 && (!bytes.Equal(input.Value, trace.Deliveries[from].Actual) || input.ProducerID != prior.Plan.Activities[from].ID) {
					t.Fatal("port did not receive its actual producer value")
				}
			}
		}
	}
	if ports != 63 {
		t.Fatalf("observed multi-input slots=%d, want63", ports)
	}
	suite.Cases[0].Expected["Difference"] = json.RawMessage("11")
	partial, err := ExecuteComposition(context.Background(), "joins.gooo", source, prior, suite, nativeTool())
	if err != nil || partial.FinitePassed != 48 || partial.FiniteTotal != 49 {
		t.Fatalf("partial score: %v %+v", err, partial)
	}
}

func TestCompositionInputDefectsPrecedeOptionalModelLoad(t *testing.T) {
	source, suite := inputJoinFixture(t)
	for _, change := range []string{"missing", "override", "unnamed", "wrong-type", "duplicate", "type-edge", "input-port", "arity", "parameter-mutation"} {
		t.Run(change, func(t *testing.T) {
			raw, _ := json.Marshal(suite)
			cases, _ := DecodeCompositionCases(raw)
			candidate := string(source)
			switch change {
			case "missing":
				delete(cases.Cases[0].Inputs, "Label.input2")
			case "override":
				cases.Cases[0].Inputs["Add.input0"] = json.RawMessage("1")
			case "unnamed":
				cases.Cases[0].Inputs["Compare"] = json.RawMessage(`"x"`)
			case "wrong-type":
				cases.Cases[0].Inputs["Label.input1"] = json.RawMessage("0")
			case "duplicate":
				candidate += "\nbind Right.result -> Add.input0\n"
			case "type-edge":
				candidate += "\nbind Compare.result -> Label.input2\n"
			case "input-port":
				candidate = strings.Replace(candidate, "Add.input0", "Add.input", 1)
			case "arity":
				candidate = strings.Replace(candidate, "Add(Integer, Integer)", "Add("+strings.TrimSuffix(strings.Repeat("Integer, ", 17), ", ")+")", 1)
			case "parameter-mutation":
				candidate = strings.Replace(candidate, "return input0 + input1", "input1 = 1; return input0", 1)
			}
			result, err := GenerateComposition(context.Background(), "invalid.gooo", []byte(candidate), cases, "missing-model.json")
			if err == nil || result.Model != nil || len(result.Steps) != 0 {
				t.Fatalf("input defect reached model/generation: %v", err)
			}
		})
	}
}

func TestCompositionFullSixteenInputArrayAndReplayMetadata(t *testing.T) {
	inputs := strings.TrimSuffix(strings.Repeat("Integer, ", 16), ", ")
	source := []byte(fmt.Sprintf("package wide\nnamespace wide\nentity Integer id \"wide://integer\"\nactivity Start(%s) -> Integer computes \"return input0 - input15\"\nactivity End(Integer) -> Integer computes \"return input\"\nbind Start.result -> End.input\n", inputs))
	test := CompositionCase{Inputs: map[string]json.RawMessage{}, Expected: map[string]json.RawMessage{"End": json.RawMessage("-15")}}
	for i := range 16 {
		test.Inputs[fmt.Sprintf("Start.input%d", i)] = json.RawMessage(fmt.Sprint(i))
	}
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1", Cases: []CompositionCase{test}}
	prior, err := GenerateComposition(context.Background(), "wide.gooo", source, suite, "")
	if err != nil {
		t.Fatal(err)
	}
	run, err := ExecuteComposition(context.Background(), "wide.gooo", source, prior, suite, nativeTool())
	if err != nil || run.FinitePassed != 1 || run.FiniteTotal != 1 {
		t.Fatalf("full parameter array: %v %+v", err, run)
	}
	for i := range prior.Steps {
		if len(prior.Steps[i].Generation.Report.InputParameters) > 0 {
			prior.Steps[i].Generation.Report.InputParameters[0].Name = "changed"
		}
	}
	if err := VerifyComposition(context.Background(), "wide.gooo", source, prior); err == nil {
		t.Fatal("altered input metadata replayed")
	}
}
