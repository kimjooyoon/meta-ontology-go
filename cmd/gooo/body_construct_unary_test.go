package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestBodyConstructUnaryArithmetic(t *testing.T) {
	root := t.TempDir()
	base := "../../examples/caller-typed-paths/"
	source, err := os.ReadFile(base + "main.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	filename := filepath.Join(root, "unary.gooo")
	if err := os.WriteFile(filename, []byte(strings.Replace(string(source), "0 - input", "-input", 1)), 0600); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(root, "constructed")
	goBin := filepath.Join(runtime.GOROOT(), "bin", "go")
	args := []string{"body-construct", "--source", filename, "--entry", "Main", "--construction-cases", base + "construction-cases.json",
		"--cases", base + "evaluation-cases.json", "--attempts", "2", "--go-bin", goBin, "--out", out}
	var stdout, stderr bytes.Buffer
	if code := run(args, &stdout, &stderr); code != exitOK {
		t.Fatal(code, stderr.String())
	}
	var result bodyConstructOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil || result.Evaluation.Runtime.FinitePassed != 3 {
		t.Fatal(result, err)
	}
	stdout.Reset()
	stderr.Reset()
	args = []string{"body-construct", "--source", filepath.Join(out, "original.gooo"), "--construction", filepath.Join(out, "construction.json"),
		"--cases", base + "evaluation-cases.json", "--go-bin", goBin}
	if code := run(args, &stdout, &stderr); code != exitOK {
		t.Fatal(code, stderr.String())
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil || result.GeneratedNow || !result.Evaluation.ConstructionReplayed ||
		result.Evaluation.NewModelCalls != 0 || result.Evaluation.Runtime.FinitePassed != 3 {
		t.Fatal(result, err)
	}
}
