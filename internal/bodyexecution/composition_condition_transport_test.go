package bodyexecution

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func TestNativeConditionFileKeepsValueExecutionAtCaptureLimit1213(t *testing.T) {
	source := []byte(`package loops1213
namespace loops1213
entity Integer id "loops1213://integer"
activity Scan(Integer) -> Integer id "loops1213://scan" computes "let cursor = 0; for cursor = 0; cursor < 1024; cursor = cursor + 1 { if cursor == input { return cursor } }; return -1"
`)
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1"}
	seen := 0
	for input := 1000; input < 1080; input++ {
		want, count := -1, 1024
		if input < 1024 {
			want, count = input, input+1
		}
		seen += count
		suite.Cases = append(suite.Cases, CompositionCase{Inputs: map[string]json.RawMessage{"Scan": json.RawMessage(fmt.Sprint(input))},
			Expected: map[string]json.RawMessage{"Scan": json.RawMessage(fmt.Sprint(want))}})
	}
	ctx := context.Background()
	prior, err := GenerateComposition(ctx, "loops1213.gooo", source, suite, "")
	if err != nil {
		t.Fatal(err)
	}
	original := compositionDigest(prior)
	run, err := ExecuteComposition(ctx, "loops1213.gooo", source, prior, suite, nativeTool())
	if err != nil || run.Stage != "COMPLETE" || !run.RuntimeReplayed || run.FinitePassed != 80 || run.FiniteTotal != 80 ||
		len(run.ConditionSites) != 1 || run.ConditionObservation == nil || run.ConditionObservation.Status != "PARTIAL" ||
		run.ConditionObservation.RecordedEvents != conditionObservationLimit || run.ConditionObservation.DroppedEvents != seen-conditionObservationLimit ||
		run.ModelCalls != 0 {
		t.Fatalf("capture limit interrupted value execution: %v stage=%s summary=%+v matched=%d/%d", err, run.Stage,
			run.ConditionObservation, run.FinitePassed, run.FiniteTotal)
	}
	recorded, dropped := 0, 0
	for _, trace := range run.Traces {
		if !trace.ConditionsObserved {
			t.Fatal("partial capture lost its observed status")
		}
		recorded += len(trace.Conditions)
		dropped += trace.ConditionsDropped
	}
	if recorded != conditionObservationLimit || recorded+dropped != seen || compositionDigest(prior) != original {
		t.Fatal("recorded and dropped denominators or saved source differ", recorded, dropped, seen)
	}
	for _, process := range run.Runs {
		if !process.Completed || process.OutputTruncated || process.ExitCode == nil || *process.ExitCode != 0 {
			t.Fatal("diagnostics overflowed the native value channel", process)
		}
	}
	changed := run
	copy := *run.ConditionObservation
	changed.ConditionObservation = &copy
	changed.ConditionObservation.DroppedEvents++
	if sameJointRuntime(run, changed) {
		t.Fatal("changed partial capture summary accepted")
	}
	saveTransportObservation(t, "loops1213", source, run)
	t.Logf("loops1213 native80/80, observedif%d recorded%d dropped%d, partial explicit, stdout complete, native file and values agree, source unchanged/model0",
		seen, recorded, dropped)
}

func TestNativeConditionFileRetainsCompleteSmallPath1247(t *testing.T) {
	source := []byte(`package thresholds1247
namespace thresholds1247
entity Integer id "thresholds1247://integer"
activity Evaluate(Integer) -> Integer id "thresholds1247://evaluate" computes "if input < 37 { return input + 3 }; return input - 2"
`)
	suite := calledCompositionCases(t, `{"schema":"gooo/body-composition-cases/v1","cases":[
{"inputs":{"Evaluate":15},"expected":{"Evaluate":18}},
{"inputs":{"Evaluate":43},"expected":{"Evaluate":41}}
]}`)
	ctx := context.Background()
	prior, err := GenerateComposition(ctx, "thresholds1247.gooo", source, suite, "")
	if err != nil {
		t.Fatal(err)
	}
	run, err := ExecuteComposition(ctx, "thresholds1247.gooo", source, prior, suite, nativeTool())
	if err != nil || run.FinitePassed != 2 || !run.RuntimeReplayed || run.ConditionObservation == nil ||
		run.ConditionObservation.Status != "COMPLETE" || run.ConditionObservation.RecordedEvents != 2 ||
		run.ConditionObservation.DroppedEvents != 0 || len(run.Traces[0].Conditions) != 1 ||
		!run.Traces[0].Conditions[0].Value || len(run.Traces[1].Conditions) != 1 || run.Traces[1].Conditions[0].Value {
		t.Fatal("small actual path was lost", err, run.ConditionObservation)
	}
	saveTransportObservation(t, "thresholds1247", source, run)
	t.Log("thresholds1247 native2/2, if events2, complete, two outputs and evidence files agree/model0")
}

func saveTransportObservation(t *testing.T, name string, source []byte, run CompositionRuntime) {
	t.Helper()
	if root := os.Getenv("GOOO_TRANSPORT_OBSERVATION_OUTPUT"); root != "" {
		value := map[string]any{"source": string(source), "source_sha256": digest(source), "stage": run.Stage,
			"finite_passed": run.FinitePassed, "finite_total": run.FiniteTotal, "condition_sites": run.ConditionSites,
			"condition_observation": run.ConditionObservation, "runtime_replayed": run.RuntimeReplayed,
			"generated_sha256": run.GeneratedSHA256, "observed_projection_sha256": run.ObservedProjectionSHA256,
			"observed_driver_sha256": run.ObservedDriverSHA256, "model_calls": run.ModelCalls, "runs": run.Runs}
		if run.ConditionObservation.DroppedEvents == 0 {
			value["traces"] = run.Traces
		}
		raw, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		file, err := os.OpenFile(root+"/"+name+".json", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			t.Fatal(err)
		}
		_, writeErr := file.Write(append(raw, '\n'))
		closeErr := file.Close()
		if writeErr != nil || closeErr != nil {
			t.Fatal(writeErr, closeErr)
		}
	}
}
