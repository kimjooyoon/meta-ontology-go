package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestAuthoredCompoundExampleNative107513(t *testing.T) {
	path := "../../examples/compound-updates/main.gooo"
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	cases := CompositionCases{Schema: "gooo/body-composition-cases/v1"}
	for _, row := range [][4]string{{"8", "6", "true", "14"}, {"19", "4", "false", "10"}, {"-13", "-4", "true", "-17"}, {"-4", "7", "false", "0"}} {
		cases.Cases = append(cases.Cases, CompositionCase{Inputs: map[string]json.RawMessage{"Adjust.input0": json.RawMessage(row[0]), "Adjust.input1": json.RawMessage(row[1]), "Adjust.input2": json.RawMessage(row[2])}, Expected: map[string]json.RawMessage{"Adjust": json.RawMessage(row[3])}})
	}
	observation := map[string]any{"source": string(source), "cases": cases}
	defer saveReadObservation(t, "compound-example107513.json", observation)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	program, err := GenerateCompositionWithOptions(ctx, path, source, cases, CompositionOptions{EntryActivity: "Adjust"})
	observation["composition"] = program
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := ExecuteComposition(ctx, path, source, program, cases, "")
	observation["runtime"] = runtime
	if err != nil || runtime.FinitePassed != 4 || !runtime.RuntimeReplayed || runtime.ModelCalls != 0 {
		t.Fatal("source example behavior", runtime.FinitePassed, err)
	}
	if runtime.LocalObservation == nil || runtime.LocalObservation.RecordedWrites != 16 || runtime.LocalObservation.RecordedUses != 16 {
		t.Fatal("source update origins", runtime.LocalObservation)
	}
	for _, trace := range runtime.Traces {
		assertCompoundTrace79243(t, trace, runtime.LocalSites, runtime.LocalUseSites)
	}
	if err := validateRecordedLocalOrigins(runtime); err != nil {
		t.Fatal(err)
	}
	t.Log("published source:new four inputs pass; sixteen writes and reads retain source; no model calls")
}
