package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestCompositionSelectionInputsIncludeSourceHoldout(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/ir-search-source.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	graph := compositionGraph{count: 1}
	graph.nodes[0] = CompositionActivity{Name: "ClampNegativeToZero", InputType: "Integer", Assembling: true}
	prior := Composition{Steps: make([]CompositionStep, 1)}
	seen, err := graph.selectionInputs(context.Background(), "search.gooo", source, prior)
	if err != nil {
		t.Fatal(err)
	}
	for input, want := range map[string]bool{"-2": true, "9223372036854775807": true,
		"9007199254740992": false, "9007199254740993": false} {
		key, err := graph.selectionInputKey(graph.nodes[0], []json.RawMessage{json.RawMessage(input)})
		_, found := seen[0][key]
		if err != nil || found != want {
			t.Fatalf("source selection/holdout identity %s: %v %v", input, found, err)
		}
	}
}

func TestRecordedSelectionInputsIncludeProbesAndOracleFeedback(t *testing.T) {
	report := bodycodegen.Report{
		BodyFill: &bodycodegen.IRBodyFillReceipt{
			SelectedCaseResults: []bodycodegen.IRBodyFillCaseResult{{Input: 1}},
			HoldoutCaseResults:  []bodycodegen.IRBodyFillCaseResult{{Input: 2}},
			BehavioralProbes:    &bodycodegen.IRBodyFillBehavioralProbeReceipt{ProbeInputs: []int64{3}, ProbeVectors: [][]int64{{4, 5}}},
		},
		BodyPaths: &bodycodegen.BodyPathReceipt{NativeCases: []bodycodegen.IRBodyFillCaseResult{{Input: 6}},
			Observation: &bodycodegen.PathObservationReceipt{
				InitialCases: []pathplan.TestCase{{Input: 7}},
				Options:      bodycodegen.PathObservationOptions{Inputs: []int64{8}},
				Rounds:       []bodycodegen.PathObservationRound{{Observation: &pathplan.TestCase{Input: 9}}},
			}},
	}
	keys := map[string]bool{}
	for _, tuple := range recordedInputTuples(report) {
		raw, err := json.Marshal(tuple)
		if err != nil {
			t.Fatal(err)
		}
		keys[string(raw)] = true
	}
	for _, want := range []string{"[1]", "[2]", "[3]", "[4,5]", "[6]", "[7]", "[8]", "[9]"} {
		if !keys[want] {
			t.Fatal("previously observed input missing", want)
		}
	}
}

func TestCompositionInputSeparationWithNoExecutedOrKnownSelection(t *testing.T) {
	if got := initialCompositionRuntime(nil, Composition{}, CompositionCases{}).InputSeparation; got.Status != "UNKNOWN" {
		t.Fatal("unexecuted composition received input coverage", got)
	}
	source, suite, prior := retainedFieldFixture(t)
	graph, err := prepareCompositionGraph(context.Background(), "fields.gooo", source)
	if err != nil {
		t.Fatal(err)
	}
	prior.Steps[0].Generation.Report.BodySearch = &bodycodegen.IRBodySearchReceipt{
		ExternalTrainingFeedback: &bodycodegen.IRBodySearchExternalFeedbackReceipt{ObservationCount: 1}}
	traces := make([]CompositionTrace, len(suite.Cases))
	got := graph.measureInputSeparation(context.Background(), "fields.gooo", source, prior, suite, traces)
	if got.Status != "UNKNOWN" || got.Reason != "SELECTION_INPUTS_UNAVAILABLE" || got.DisjointInputs != 0 {
		t.Fatal("unavailable external input identities gained coverage", got)
	}
}
