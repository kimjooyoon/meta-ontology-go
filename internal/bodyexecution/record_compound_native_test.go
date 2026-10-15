package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"testing"
)

func TestRecordCompoundNativeOrigins(t *testing.T) {
	const path = "../../examples/compound-updates/fields.gooo"
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	cases := recordCompoundNativeCases()
	observation := map[string]any{"source": string(source), "cases": cases}
	defer saveReadObservation(t, "record-compound-native.json", observation)
	ctx := context.Background()
	program, err := GenerateCompositionWithOptions(ctx, path, source, cases, CompositionOptions{EntryActivity: "Summarize"})
	observation["composition"] = program
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := ExecuteComposition(ctx, path, source, program, cases, "")
	observation["runtime"] = runtime
	if err != nil || runtime.FinitePassed != 4 || runtime.ModelCalls != 0 || !runtime.RuntimeReplayed {
		t.Fatal("native record updates", err, runtime.FinitePassed)
	}
	if runtime.FieldObservation == nil || runtime.FieldObservation.Status != "COMPLETE" ||
		runtime.BridgeObservation == nil || runtime.BridgeObservation.Status != "COMPLETE" {
		t.Fatal("record/local origin coverage", runtime.FieldObservation, runtime.BridgeObservation)
	}
	assertRecordCompoundJournal(t, runtime)
	if err := validateFieldSources(runtime); err != nil {
		t.Fatal(err)
	}
	t.Log("four new native cases; sixteen compound field writes retain preceding field reads, scalar-local operands and source; model calls 0")
}

func recordCompoundNativeCases() CompositionCases {
	suite := CompositionCases{Schema: "gooo/body-composition-cases/v1"}
	for _, row := range [][3]string{
		{`{"count":8,"step":3,"title":"a","untouched":11}`, "true", `{"count":8,"step":3,"title":"a✓","untouched":11}`},
		{`{"count":5,"step":2,"title":"한글","untouched":13}`, "false", `{"count":1,"step":2,"title":"한글✓","untouched":13}`},
		{`{"count":-7,"step":4,"title":"","untouched":17}`, "true", `{"count":-1,"step":4,"title":"✓","untouched":17}`},
		{`{"count":0,"step":-2,"title":"x","untouched":19}`, "false", `{"count":0,"step":-2,"title":"x✓","untouched":19}`},
	} {
		suite.Cases = append(suite.Cases, CompositionCase{Inputs: map[string]json.RawMessage{
			"Summarize.input0": json.RawMessage(row[0]), "Summarize.input1": json.RawMessage(row[1])},
			Expected: map[string]json.RawMessage{"Summarize": json.RawMessage(row[2])}})
	}
	return suite
}
