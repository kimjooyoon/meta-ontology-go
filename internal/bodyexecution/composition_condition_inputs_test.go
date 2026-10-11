package bodyexecution

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

const conditionInputSource = `package conditioninputs
namespace conditioninputs
entity Integer id "conditioninputs://integer"
activity Choose(Integer) -> Integer computes "if input < 5 { return input } else { return input + 1 }" assembling {
    choice "boundary" operand_order at "0" intent "Check the input against five."
    case "0" -> "0"
    condition_case "boundary" input "7" -> "false"
    condition_case "boundary" input "9007199254740993" -> "false"
    condition_case "boundary" input "-9007199254740993" -> "true"
    attempts "2"
}
`

func TestCompositionSelectionInputsIncludeDeclaredConditions(t *testing.T) {
	graph := compositionGraph{count: 1}
	graph.nodes[0] = CompositionActivity{Name: "Choose", InputType: "Integer", Assembling: true}
	prior := Composition{Steps: make([]CompositionStep, 1)}
	seen, err := graph.selectionInputs(context.Background(), "condition-inputs.gooo", []byte(conditionInputSource), prior)
	if err != nil {
		t.Fatal(err)
	}
	for input, want := range map[string]bool{"0": true, "7": true, "9": false,
		"9007199254740993": true, "-9007199254740993": true, "9007199254740992": false} {
		key, err := graph.selectionInputKey(graph.nodes[0], []json.RawMessage{json.RawMessage(input)})
		_, found := seen.nodes[0][key]
		if err != nil || found != want {
			t.Fatalf("condition-only input %s: found %v, want %v, error %v", input, found, want, err)
		}
	}
}

func TestCompositionConditionOnlyInputsRemainOverlappingInNativeCalls(t *testing.T) {
	for _, entry := range []string{"Choose", "Main"} {
		t.Run(entry, func(t *testing.T) {
			source, shift := conditionInputSource, int64(0)
			if entry == "Main" {
				source += "activity Main(Integer) -> Integer computes `return Choose(input - 2)`\n"
				shift = 2
			}
			suite := CompositionCases{Schema: "gooo/body-composition-cases/v1"}
			for _, input := range []int64{7, 9, 9007199254740993, -9007199254740993} {
				expected := input
				if input >= 5 {
					expected++
				}
				suite.Cases = append(suite.Cases, CompositionCase{
					Inputs:   map[string]json.RawMessage{entry: json.RawMessage(fmt.Sprint(input + shift))},
					Expected: map[string]json.RawMessage{entry: json.RawMessage(fmt.Sprint(expected))}})
			}
			ctx := context.Background()
			prior, err := GenerateCompositionWithOptions(ctx, "condition-inputs.gooo", []byte(source), suite,
				CompositionOptions{EntryActivity: entry})
			if err != nil {
				t.Fatal(err)
			}
			runtime, err := ExecuteComposition(ctx, "condition-inputs.gooo", []byte(source), prior, suite, nativeTool())
			if err != nil {
				t.Fatal(err)
			}
			got := runtime.InputSeparation
			if runtime.FinitePassed != 4 || got.Status != "PASS" || got.UniqueInputs != 4 ||
				got.OverlappingInputs != 3 || got.DisjointInputs != 1 || got.DisjointCasesPassed != 1 || got.UnknownInputs != 0 {
				t.Fatal("condition-only inputs inflated disjoint coverage", got)
			}
			if entry == "Main" && (!runtime.Traces[0].CalledInputsObserved || len(runtime.Traces[0].Calls) != 1 ||
				string(runtime.Traces[0].Calls[0].Inputs[0]) != "7") {
				t.Fatal("root input was substituted for the actual helper input", runtime.Traces[0])
			}
		})
	}
}
