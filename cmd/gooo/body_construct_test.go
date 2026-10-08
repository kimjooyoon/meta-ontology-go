package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestBodyConstructCallerFeedbackAndSavedReplay(t *testing.T) {
	root := filepath.Join(t.TempDir(), "constructed")
	base := "../../examples/caller-guided-construction/"
	args := []string{"body-construct", "--source", base + "main.gooo.fixture", "--entry", "Main",
		"--construction-cases", base + "construction-cases.json", "--cases", base + "evaluation-cases.json",
		"--attempts", "2", "--go-bin", filepath.Join(runtime.GOROOT(), "bin", "go"), "--out", root}
	var stdout, stderr bytes.Buffer
	if code := run(args, &stdout, &stderr); code != exitOK {
		t.Fatal(code, stderr.String())
	}
	var result bodyConstructOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if !result.GeneratedNow || result.Construction.Decision != "COMPLETE_FINITE" || result.Evaluation.Runtime.FinitePassed != 4 ||
		result.Evaluation.InputSeparation.ConstructionInputs != 1 || result.Evaluation.InputSeparation.OtherInputs != 3 || result.Evaluation.ConstructionReplayed {
		t.Fatal("construction/evaluation separation differs", result)
	}
	for _, name := range []string{"original.gooo", "selected.gooo", "construction.json", "evaluation.json", "generated.go", "main.go"} {
		if _, err := os.Stat(filepath.Join(root, name)); err != nil {
			t.Fatal(err)
		}
	}
	stdout.Reset()
	stderr.Reset()
	args = []string{"body-construct", "--source", filepath.Join(root, "original.gooo"), "--construction", filepath.Join(root, "construction.json"),
		"--cases", base + "evaluation-cases.json", "--go-bin", filepath.Join(runtime.GOROOT(), "bin", "go")}
	if code := run(args, &stdout, &stderr); code != exitOK {
		t.Fatal(code, stderr.String())
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.GeneratedNow || !result.Evaluation.ConstructionReplayed || result.Evaluation.NewModelCalls != 0 || result.Evaluation.Runtime.FinitePassed != 4 {
		t.Fatal("saved construction did not replay", result.Evaluation)
	}
}

func TestBodyConstructRejectsAmbiguousModes(t *testing.T) {
	base := []string{"--source", "source", "--cases", "cases"}
	for _, flags := range [][]string{{}, {"--attempts", "0", "--construction-cases", "feedback"},
		{"--attempts", "65", "--construction-cases", "feedback"},
		{"--construction", "saved", "--model", "model"}, {"--construction", "saved", "--entry", "Main"},
		{"--construction", "saved", "--fill-model", "model"},
		{"--construction", "saved", "--construction-cases", "feedback"}, {"--unknown", "value"}} {
		if _, err := parseBodyConstruct(append(append([]string{}, base...), flags...)); err == nil {
			t.Fatal(flags)
		}
	}
}

func TestBodyConstructSourceFillAndModelFlag(t *testing.T) {
	base := "../../examples/caller-source-fill/"
	args := []string{"body-construct", "--source", base + "budget.gooo.fixture", "--entry", "Main",
		"--construction-cases", base + "construction-cases.json", "--cases", base + "evaluation-cases.json",
		"--attempts", "3", "--go-bin", filepath.Join(runtime.GOROOT(), "bin", "go")}
	var stdout, stderr bytes.Buffer
	if code := run(args, &stdout, &stderr); code != exitOK {
		t.Fatal(code, stderr.String())
	}
	var result bodyConstructOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Construction.Schema != "gooo/joint-construction/v4" || result.Evaluation.Runtime.FinitePassed != 4 {
		t.Fatal(result)
	}
	stdout.Reset()
	stderr.Reset()
	if code := run(append(args, "--fill-model", "missing.json"), &stdout, &stderr); code != exitFailure {
		t.Fatal("fill model flag ignored", code)
	}
}
