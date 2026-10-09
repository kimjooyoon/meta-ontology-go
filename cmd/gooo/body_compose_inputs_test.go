package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

const compositionInputsFixtures = "../../examples/composition-inputs/"

func runCompositionInputs(t *testing.T, source, input string, options ...string) bodyCompositionOutput {
	t.Helper()
	args := []string{"body-compose", "--source", source, "--inputs", input,
		"--go-bin", filepath.Join(runtime.GOROOT(), "bin", "go")}
	var out, diagnostics bytes.Buffer
	if code := run(append(args, options...), &out, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String())
	}
	var result bodyCompositionOutput
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	var envelope struct {
		InputSchema string `json:"input_schema"`
	}
	if err := json.Unmarshal(out.Bytes(), &envelope); err != nil || envelope.InputSchema != bodyexecution.CompositionInputsSchema {
		t.Fatal("input mode missing from output", err, envelope)
	}
	return result
}

func requireUnscoredComposition(t *testing.T, value bodyexecution.CompositionRuntime) {
	t.Helper()
	if !value.RuntimeReplayed || value.ModelCalls != 0 || value.FinitePassed != 0 || value.FiniteTotal != 0 ||
		value.InputSeparation.Status != "UNKNOWN" || value.InputSeparation.Reason != "NO_RUNTIME_EXPECTATIONS" {
		t.Fatal("observed values acquired a correctness score", value)
	}
	for _, trace := range value.Traces {
		for _, delivery := range trace.Deliveries {
			if len(delivery.Expected) != 0 || delivery.Passed != nil {
				t.Fatal("an oracle was invented", delivery)
			}
		}
	}
}

func TestBodyComposeInputsRepeatsReplaysAndKeepsTypedBindings(t *testing.T) {
	source := "../../examples/body-codegen/native-input-joins.gooo.fixture"
	input := compositionInputsFixtures + "joins.json"
	directory := filepath.Join(t.TempDir(), "observed")
	first := runCompositionInputs(t, source, input, "--repeat", "2", "--out", directory)
	if !first.GeneratedNow || len(first.RuntimeHistory) != 2 {
		t.Fatal("input-only history missing", first)
	}
	for _, value := range first.RuntimeHistory {
		requireUnscoredComposition(t, value)
	}
	if !first.RuntimeHistory[1].Artifact.Reused || first.RuntimeHistory[1].Build.Started {
		t.Fatal("retained executable not reused", first.RuntimeHistory[1])
	}
	requireInputFiles(t, directory, input)
	trace := first.Runtime.Traces[0].Deliveries
	byID := make(map[string]bodyexecution.CompositionDelivery, len(trace))
	for _, delivery := range trace {
		byID[delivery.ActivityID] = delivery
	}
	left, add := byID["joins://activity/left"], byID["joins://activity/add"]
	if len(trace) != 7 || string(left.Input) != "9007199254740993" || string(left.Actual) != "-9007199254740986" ||
		len(add.Inputs) != 2 || add.Inputs[0].ProducerID != left.ActivityID || string(add.Inputs[0].Value) != string(left.Actual) ||
		string(add.Inputs[1].Value) != "4" {
		t.Fatal("root and intermediate inputs were mixed", trace)
	}
	replay := runCompositionInputs(t, source, input, "--composition", filepath.Join(directory, "composition.json"))
	requireUnscoredComposition(t, replay.Runtime)
	if replay.GeneratedNow || replay.Composition.GeneratedSHA256 != first.Composition.GeneratedSHA256 {
		t.Fatal("saved graph was regenerated", replay)
	}
	requireScoredReplay(t, source, directory)
}

func requireInputFiles(t *testing.T, directory, input string) {
	t.Helper()
	got, err := os.ReadFile(filepath.Join(directory, "inputs.json"))
	want, readErr := os.ReadFile(input)
	if err != nil || readErr != nil || !bytes.Equal(got, want) {
		t.Fatal("input document changed", err, readErr)
	}
	if _, err := os.Stat(filepath.Join(directory, "cases.json")); !os.IsNotExist(err) {
		t.Fatal("input document mislabeled as cases", err)
	}
}

func requireScoredReplay(t *testing.T, source, directory string) {
	t.Helper()
	var out, diagnostics bytes.Buffer
	args := []string{"body-compose", "--source", source, "--cases",
		"../../examples/body-codegen/native-input-joins-cases.json", "--composition", filepath.Join(directory, "composition.json")}
	if code := run(args, &out, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String())
	}
	var result bodyCompositionOutput
	if err := json.Unmarshal(out.Bytes(), &result); err != nil || result.GeneratedNow || result.Runtime.FinitePassed != 49 ||
		result.Runtime.FiniteTotal != 49 || result.Runtime.ModelCalls != 0 || result.InputSchema != "gooo/body-composition-cases/v1" {
		t.Fatal("later supplied expectations did not score the same saved graph", err, result.Runtime)
	}
}

func TestBodyComposeInputsRejectsInvalidRootsBeforeModelLoad(t *testing.T) {
	valid, err := os.ReadFile(compositionInputsFixtures + "record.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, change := range [][2]string{{"Describe.input0", "Unknown.input0"},
		{"9007199254740993", `"9007199254740993"`}, {"9007199254740993", "null"},
		{`"Describe.input1":true,`, ""}, {`"inputs":`, `"expected":{},"inputs":`},
		{"inputs/v1", "cases/v1"}} {
		t.Run(change[1], func(t *testing.T) {
			input := filepath.Join(t.TempDir(), "inputs.json")
			if err := os.WriteFile(input, []byte(strings.Replace(string(valid), change[0], change[1], 1)), 0600); err != nil {
				t.Fatal(err)
			}
			var out, diagnostics bytes.Buffer
			args := []string{"body-compose", "--source", "../../examples/scalar-identity/source.gooo.fixture",
				"--entry", "Describe", "--inputs", input, "--model", "model-must-not-be-loaded.json"}
			if code := run(args, &out, &diagnostics); code != exitFailure ||
				strings.Contains(diagnostics.String(), "model-must-not-be-loaded") {
				t.Fatal("invalid input reached the model", code, diagnostics.String())
			}
		})
	}
}

func TestBodyComposeInputsCannotReplaceBoundProducerValue(t *testing.T) {
	valid, err := os.ReadFile(compositionInputsFixtures + "joins.json")
	if err != nil {
		t.Fatal(err)
	}
	input := filepath.Join(t.TempDir(), "inputs.json")
	if err := os.WriteFile(input, []byte(strings.Replace(string(valid), `"Right":2`, `"Right":2,"Add.input0":1`, 1)), 0600); err != nil {
		t.Fatal(err)
	}
	var out, diagnostics bytes.Buffer
	args := []string{"body-compose", "--source", "../../examples/body-codegen/native-input-joins.gooo.fixture", "--inputs", input}
	if code := run(args, &out, &diagnostics); code != exitFailure || !strings.Contains(diagnostics.String(), "bound input") {
		t.Fatal("caller replaced a producer", code, diagnostics.String())
	}
}

func TestBodyComposeInputsOwnModelPreservesExactRecordValues(t *testing.T) {
	source := "../../examples/scalar-identity/source.gooo.fixture"
	input := compositionInputsFixtures + "record.json"
	directory := filepath.Join(t.TempDir(), "observed")
	first := runCompositionInputs(t, source, input, "--entry", "Describe", "--model",
		"../../examples/scalar-identity/model/model.json", "--out", directory)
	requireUnscoredComposition(t, first.Runtime)
	assembly := first.Composition.Steps[0].Generation.Report.RecordAssembly
	if assembly.ModelCalls != 1 || assembly.Passed != assembly.Total {
		t.Fatal("source-owned model assembly changed", assembly)
	}
	var actual map[string]json.RawMessage
	if err := json.Unmarshal(first.Runtime.Traces[0].Deliveries[0].Actual, &actual); err != nil ||
		string(actual["count"]) != "9007199254740993" || string(actual["text"]) != `"해보자!"` || string(actual["active"]) != "true" {
		t.Fatal("exact record observation lost", err, actual)
	}
	if err := json.Unmarshal(first.Runtime.Traces[1].Deliveries[0].Actual, &actual); err != nil ||
		string(actual["count"]) != "0" || string(actual["text"]) != `""` || string(actual["active"]) != "false" {
		t.Fatal("zero values became absent", err, actual)
	}
	replay := runCompositionInputs(t, source, input, "--composition", filepath.Join(directory, "composition.json"))
	requireUnscoredComposition(t, replay.Runtime)
	if replay.GeneratedNow || replay.Composition.Steps[0].Generation.Report.RecordAssembly.ModelCalls != 1 {
		t.Fatal("original inference was overwritten", replay)
	}
}

func TestBodyComposeInputsResumeNestedHelpers(t *testing.T) {
	source := "../../examples/dependent-continuation/main.gooo.fixture"
	input := compositionInputsFixtures + "helpers.json"
	directory := filepath.Join(t.TempDir(), "partial")
	first := runCompositionInputs(t, source, input, "--entry", "Main", "--assembly-policy",
		"../../examples/assembly-policy/checkpoint.gooo.fixture", "--policy-activity", "Checkpoint", "--out", directory)
	requireUnscoredComposition(t, first.Runtime)
	resumed := runCompositionInputs(t, source, input, "--resume-composition", filepath.Join(directory, "composition.json"),
		"--assembly-policy", "../../examples/assembly-explainer/main.gooo.fixture", "--policy-activity", "Explain")
	requireUnscoredComposition(t, resumed.Runtime)
	if !resumed.GeneratedNow || resumed.Composition.Continuation.NewModelCalls != 0 ||
		len(resumed.Runtime.Traces[0].Calls) != 4 || resumed.Composition.Continuation.Activities[1].RecheckedAttempts != 1 {
		t.Fatal("input-only continuation changed dependencies or inferred again", resumed)
	}
}

func TestBodyComposeInputsModesAndSchemasStaySeparate(t *testing.T) {
	base := []string{"body-compose", "--source", "unused", "--inputs", "unused"}
	for _, extra := range [][]string{{"--cases", "unused"}, {"--case-series", "unused"},
		{"--cases", "unused", "--case-series", "unused"}} {
		var out, diagnostics bytes.Buffer
		if code := run(append(append([]string(nil), base...), extra...), &out, &diagnostics); code != exitUsage || out.Len() != 0 {
			t.Fatal("ambiguous input mode", code, diagnostics.String())
		}
	}
	for _, args := range [][]string{{"help", "body-compose"}, {"body-compose", "--help"}} {
		var out, diagnostics bytes.Buffer
		if code := run(args, &out, &diagnostics); code != exitOK || !strings.Contains(out.String(), "--inputs") ||
			!strings.Contains(out.String(), "NO_RUNTIME_EXPECTATIONS") {
			t.Fatal("input-only guide missing", code, out.String(), diagnostics.String())
		}
	}
}
