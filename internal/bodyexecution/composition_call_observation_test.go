package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestCalledInputObservationFollowsArgumentsAndBranches(t *testing.T) {
	source := []byte(calledRecordSeed + `
activity Main(Integer) -> Integer computes "if input < 0 { return 0 }; let result = Seed(input - 100); return result.value"
`)
	suite := calledCompositionCases(t, `{"schema":"gooo/body-composition-cases/v1","cases":[
{"inputs":{"Main":101},"expected":{"Main":2}},
{"inputs":{"Main":103},"expected":{"Main":4}},
{"inputs":{"Main":103},"expected":{"Main":6}},
{"inputs":{"Main":-1},"expected":{"Main":0}}
]}`)
	ctx := context.Background()
	prior, err := GenerateCompositionWithOptions(ctx, "arguments.gooo", source, suite, CompositionOptions{EntryActivity: "Main"})
	if err != nil {
		t.Fatal(err)
	}
	native, err := ExecuteComposition(ctx, "arguments.gooo", source, prior, suite, nativeTool())
	if err != nil {
		t.Fatal(err)
	}
	got := native.InputSeparation
	if got.UniqueInputs != 3 || got.DuplicateRows != 1 || got.OverlappingInputs != 1 || got.DisjointInputs != 1 ||
		got.DisjointCasesPassed != 0 || got.UnknownInputs != 1 || got.Status != "UNKNOWN" || native.FinitePassed != 3 {
		t.Fatal("call arguments, duplicate expectations or skipped assembly misclassified", got, native.FinitePassed)
	}
	for i, expected := range []string{"1", "3", "3"} {
		trace := native.Traces[i]
		if !trace.CalledInputsObserved || len(trace.Calls) != 1 || string(trace.Calls[0].Inputs[0]) != expected ||
			trace.Calls[0].ActivityID != prior.Plan.Preparations[0].ID || trace.Calls[0].RootActivityID != prior.Plan.Activities[0].ID {
			t.Fatal("actual helper invocation was not retained", trace)
		}
	}
	if !native.Traces[3].CalledInputsObserved || len(native.Traces[3].Calls) != 0 {
		t.Fatal("skipped call invented an input")
	}
	projection, driver, err := observedCompositionSources(prior)
	if err != nil || native.ObservedProjectionSHA256 != digest([]byte(projection)) || native.ObservedDriverSHA256 != digest([]byte(driver)) ||
		strings.Contains(prior.Source, "GoooObserveCalledInputs") || VerifyComposition(ctx, "arguments.gooo", source, prior) != nil {
		t.Fatal("observation changed the saved construction or lost source identity", err)
	}
}

func TestCalledInputObservationKeepsIndirectSelectionExposureUnknown(t *testing.T) {
	source := []byte(calledRecordSeed + `
activity Wrap(Integer) -> Box computes "let next = Seed(input + 10); return Box{value: next.value}" assembling {
    choice "value" field_value at "0" alternative "next.value * 2" intent "Double."
    value_case "[1]" -> "{\"value\":24}"
    attempts "2"
}
activity Main(Integer) -> Integer computes "let result = Wrap(input); return result.value"
`)
	suite := calledCompositionCases(t, `{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"Main":3},"expected":{"Main":28}}]}`)
	ctx := context.Background()
	prior, err := GenerateCompositionWithOptions(ctx, "indirect.gooo", source, suite, CompositionOptions{EntryActivity: "Main"})
	if err != nil {
		t.Fatal(err)
	}
	native, err := ExecuteComposition(ctx, "indirect.gooo", source, prior, suite, nativeTool())
	if err != nil || native.FinitePassed != 1 || native.InputSeparation.Status != "UNKNOWN" ||
		native.InputSeparation.Reason != "CONSTRUCTION_CALL_INPUTS_NOT_OBSERVED" || native.InputSeparation.DisjointInputs != 0 {
		t.Fatal("unrecorded calls during candidate scoring became new inputs", err, native)
	}
	calls := native.Traces[0].Calls
	if len(calls) != 2 || string(calls[0].Inputs[0]) != "3" || string(calls[1].Inputs[0]) != "13" {
		t.Fatal("nested actual arguments lost", calls)
	}
}

func TestCalledInputObservationRejectsIncompleteOrWrongTypedTrace(t *testing.T) {
	source := []byte(calledRecordSeed + "activity Main(Integer) -> Box computes `return Seed(input)`\n")
	graph, err := prepareCompositionGraphForEntry(context.Background(), "trace.gooo", source, "Main")
	if err != nil {
		t.Fatal(err)
	}
	root, helper := graph.nodes[0].ID, graph.plan.Preparations[0].ID
	valid := map[string]any{"schema": "gooo/called-input-observation/v1", "outputs": [][]any{{map[string]int{"value": 4}}},
		"calls": []any{map[string]any{"case_index": 0, "root_activity_id": root, "activity_id": helper, "inputs": []any{3}}}}
	raw, _ := json.Marshal(valid)
	if _, calls, err := graph.nativeCallRows(raw, 1); err != nil || len(calls[0]) != 1 {
		t.Fatal(err, calls)
	}
	for _, call := range []map[string]any{
		{"case_index": 1, "root_activity_id": root, "activity_id": helper, "inputs": []any{3}},
		{"case_index": 0, "root_activity_id": "unknown", "activity_id": helper, "inputs": []any{3}},
		{"case_index": 0, "root_activity_id": root, "activity_id": "unknown", "inputs": []any{3}},
		{"case_index": 0, "root_activity_id": root, "activity_id": helper, "inputs": []any{"3"}},
		{"case_index": 0, "root_activity_id": root, "activity_id": helper, "inputs": []any{}},
	} {
		valid["calls"] = []any{call}
		raw, _ = json.Marshal(valid)
		if _, _, err := graph.nativeCallRows(raw, 1); err == nil {
			t.Fatal("invalid call trace accepted", call)
		}
	}
	delete(valid, "calls")
	raw, _ = json.Marshal(valid)
	if _, _, err := graph.nativeCallRows(raw, 1); err == nil {
		t.Fatal("missing trace became no calls")
	}
}

func TestCalledInputObservationRecordsTypedRecordArgumentsWithRetainedBuild(t *testing.T) {
	source := []byte(`package records
namespace records
entity Integer id "records://integer"
entity Box id "records://box" fields { field value id "records://value" type integer required one }
activity Seed(Box) -> Box computes "return Box{value: input.value}" assembling {
    choice "value" field_value at "0" alternative "input.value + 1" intent "Increment."
    value_case "[{\"value\":1}]" -> "{\"value\":2}"
    attempts "2"
}

activity Main(Integer) -> Box computes "return Seed(Box{value: input})"
`)
	suite := calledCompositionCases(t, `{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"Main":1},"expected":{"Main":{"value":2}}}]}`)
	ctx := context.Background()
	prior, err := GenerateCompositionWithOptions(ctx, "records.gooo", source, suite, CompositionOptions{EntryActivity: "Main"})
	if err != nil {
		t.Fatal(err)
	}
	executor := NewExecutor()
	defer executor.Close()
	first, err := executor.ExecuteComposition(ctx, "records.gooo", source, prior, suite, nativeTool())
	if err != nil || first.InputSeparation.OverlappingInputs != 1 {
		t.Fatal(err, first)
	}
	suite.Cases[0].Inputs["Main"] = json.RawMessage(`9007199254740993`)
	suite.Cases[0].Expected["Main"] = json.RawMessage(`{"value":9007199254740994}`)
	second, err := executor.ExecuteComposition(ctx, "records.gooo", source, prior, suite, nativeTool())
	if err != nil || !second.Artifact.Reused || second.Build.Started || second.InputSeparation.DisjointCasesPassed != 1 ||
		second.ObservedProjectionSHA256 != first.ObservedProjectionSHA256 || second.ObservedDriverSHA256 != first.ObservedDriverSHA256 ||
		string(second.Traces[0].Calls[0].Inputs[0]) != `{"value":9007199254740993}` {
		t.Fatal("retained artifact reused old inputs or lost integer precision", err, second)
	}
}

func TestCalledInputObservationPreservesRecordSourceFillProfile(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/source-ir-fill-record-composed.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	source = append(source, []byte("\nactivity Main(Candidate) -> Review computes `return ReviewCandidate(input)`\n")...)
	suite := calledCompositionCases(t, `{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"Main":{"candidate_id":45,"reviewed":true,"state":"ready"}},"expected":{"Main":{"candidate_id":45,"decision":"accepted"}}}]}`)
	ctx := context.Background()
	prior, err := GenerateCompositionWithOptions(ctx, "record-fill.gooo", source, suite, CompositionOptions{EntryActivity: "Main"})
	if err != nil {
		t.Fatal(err)
	}
	native, err := ExecuteComposition(ctx, "record-fill.gooo", source, prior, suite, nativeTool())
	if err != nil || native.FinitePassed != 1 || native.InputSeparation.DisjointCasesPassed != 1 || len(native.Traces[0].Calls) != 1 {
		t.Fatal("helper observation narrowed the source-fill profile", err, native)
	}
}
