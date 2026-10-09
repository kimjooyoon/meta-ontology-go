package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"runtime"
	"testing"
)

func TestBodyConstructTypedPathsAndSavedReplay(t *testing.T) {
	root := filepath.Join(t.TempDir(), "paths")
	base := "../../examples/caller-typed-paths/"
	args := []string{"body-construct", "--source", base + "main.gooo.fixture", "--entry", "Main",
		"--construction-cases", base + "construction-cases.json", "--cases", base + "evaluation-cases.json",
		"--attempts", "2", "--go-bin", filepath.Join(runtime.GOROOT(), "bin", "go"), "--out", root}
	var stdout, stderr bytes.Buffer
	if code := run(args, &stdout, &stderr); code != exitOK {
		t.Fatal(code, stderr.String())
	}
	var result bodyConstructOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil ||
		result.Construction.Schema != "gooo/joint-construction/v7" || result.Evaluation.Runtime.FinitePassed != 3 {
		t.Fatal("typed path CLI did not evaluate all separate inputs", err)
	}
	stdout.Reset()
	stderr.Reset()
	args = []string{"body-construct", "--source", filepath.Join(root, "original.gooo"),
		"--construction", filepath.Join(root, "construction.json"), "--cases", base + "evaluation-cases.json",
		"--go-bin", filepath.Join(runtime.GOROOT(), "bin", "go")}
	if code := run(args, &stdout, &stderr); code != exitOK {
		t.Fatal(code, stderr.String())
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil || result.GeneratedNow ||
		!result.Evaluation.ConstructionReplayed || result.Evaluation.NewModelCalls != 0 || result.Evaluation.Runtime.FinitePassed != 3 {
		t.Fatal("saved typed path CLI did not replay", err)
	}
}
