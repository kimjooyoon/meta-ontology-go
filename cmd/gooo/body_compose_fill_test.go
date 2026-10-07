package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestBodyComposeCLISourceFillRetainsModelAndReplaysWithoutIt(t *testing.T) {
	model := writeSyntheticTinyGoModel(t, "add")
	directory := filepath.Join(t.TempDir(), "composition")
	tool := filepath.Join(runtime.GOROOT(), "bin", "go")
	if runtime.GOOS == "windows" {
		tool += ".exe"
	}
	source := "../../examples/body-codegen/source-fill-composition.gooo.fixture"
	cases := "../../examples/body-codegen/source-fill-composition-cases.json"
	args := []string{"body-compose", "--source", source, "--cases", cases, "--go-bin", tool,
		"--fill-model", model, "--out", directory}
	result := runFillCompositionCLI(t, args)
	info := result.Composition.FillModel
	if info == nil || !info.Loaded || info.Loads != 1 || result.Runtime.FinitePassed != 15 {
		t.Fatal("model retention or native outputs differ", info)
	}
	for i, step := range result.Composition.Steps[:2] {
		fill := step.Generation.Report.BodyFill
		if fill == nil || fill.Decision.Provider != "tiny_go" || fill.LocalModelPredictions != 1 ||
			fill.ExternalProviderCalls != 0 || !fill.ExternalProviderCallsKnown {
			t.Fatal("local fill accounting differs", fill)
		}
		if (fill.Timing.TinyModelLoadMS != nil) != (i == 0) {
			t.Fatal("model load counted more than once")
		}
	}
	second := result.Composition.Steps[1].Generation.Report.BodyFill
	if second.ProposedCandidateID != "add_one" || second.SelectedCandidateID != "double" || second.SelectionAdjustment == "" {
		t.Fatal("wrong model proposal was not preserved beside the corrected selection", second)
	}
	if err := os.RemoveAll(filepath.Dir(model)); err != nil {
		t.Fatal(err)
	}
	// A configured external provider must never be inferred by composition replay.
	t.Setenv("GOOO_LAYA_URL", "http://127.0.0.1:1/v1/systemone")
	saved := []string{"body-compose", "--source", source, "--cases", cases, "--go-bin", tool,
		"--composition", filepath.Join(directory, "composition.json")}
	replayed := runFillCompositionCLI(t, saved)
	if replayed.GeneratedNow || replayed.Runtime.ModelCalls != 0 || replayed.Runtime.FinitePassed != 15 || !replayed.Runtime.RuntimeReplayed {
		t.Fatal("saved fill required a model", replayed.Runtime)
	}
}

func runFillCompositionCLI(t *testing.T, args []string) bodyCompositionOutput {
	t.Helper()
	var stdout, stderr bytes.Buffer
	if code := run(args, &stdout, &stderr); code != exitOK {
		t.Fatal(code, stderr.String())
	}
	var result bodyCompositionOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func TestBodyComposeCLISourceFillRejectsModelDuringReplay(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"body-compose", "--source", "unused", "--cases", "unused", "--composition", "unused",
		"--fill-model", "unused"}, &stdout, &stderr)
	if code != exitUsage {
		t.Fatal("replay accepted a new model", code, stderr.String())
	}
}

func TestBodyRealizeCLISourceFillRecordReplaysTinyDecision(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	model := writeSyntheticTinyGoModel(t, "and")
	source := "../../examples/body-codegen/source-ir-fill-record-tiny.gooo.fixture"
	var generated, stderr bytes.Buffer
	if code := run([]string{"body-codegen", "--json", "--tiny-model", model, "--activity", "ReviewCandidate", source}, &generated, &stderr); code != exitOK {
		t.Fatal(code, stderr.String())
	}
	directory := t.TempDir()
	generation := filepath.Join(directory, "generation.json")
	if err := os.WriteFile(generation, generated.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(filepath.Dir(model)); err != nil {
		t.Fatal(err)
	}
	var stdout bytes.Buffer
	if code := run([]string{"body-realize", "--source", source, "--generation", generation, "--out", filepath.Join(directory, "realized")}, &stdout, &stderr); code != exitOK {
		t.Fatal(code, stderr.String())
	}
	var realized bodycodegen.Realization
	if err := json.Unmarshal(stdout.Bytes(), &realized); err != nil || realized.FinitePassed != 2 || realized.ModelCalls != 0 {
		t.Fatal("record TinyGo decision did not replay", err, realized)
	}
}
