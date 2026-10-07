package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunPackageReplayRelocatesAndRunsNewInputsWithoutProvider(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := filepath.Join("..", "..", "examples", "package-diagnostic-replay")
	var stdout, stderr bytes.Buffer
	if code := run([]string{"package", "execute", "--json", "--cases", filepath.Join(root, "cases.json"), filepath.Join(root, "gooo.workspace.json")}, &stdout, &stderr); code != exitOK {
		t.Fatal(code, stderr.String(), stdout.String())
	}
	receiptPath := filepath.Join(t.TempDir(), "execution.json")
	if err := os.WriteFile(receiptPath, stdout.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	// Relative source identities allow moving the entire workspace to another root.
	copyRoot := t.TempDir()
	for _, name := range []string{"gooo.workspace.json", "diagnostics.gooo.fixture", "app.gooo.fixture", "inputs.json", "cases.json"} {
		raw, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(copyRoot, name), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("GOOO_LAYA_URL", "http://127.0.0.1:1/provider-must-not-be-used")
	t.Setenv("GOOO_LAYA_API_KEY", "unused")
	stdout.Reset()
	args := []string{"package", "replay", "--json", "--receipt", receiptPath, "--inputs", filepath.Join(copyRoot, "inputs.json"), filepath.Join(copyRoot, "gooo.workspace.json")}
	if code := run(args, &stdout, &stderr); code != exitOK {
		t.Fatal(code, stderr.String(), stdout.String())
	}
	var receipt packageExecutionReceipt
	if err := json.Unmarshal(stdout.Bytes(), &receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.Decision != "OBSERVED" || receipt.InputsDigest == "" || receipt.CasesDigest != "" || receipt.ReplayedFrom == "" || receipt.Result == nil {
		t.Fatal("replay envelope differs", receipt)
	}
	r := receipt.Result
	if r.Replay == nil || r.Replay.ModelCalls != 0 || r.Runtime.ModelCalls != 0 || r.Runtime.FiniteTotal != 0 || !r.Runtime.RuntimeReplayed {
		t.Fatal("replay called inference or invented correctness")
	}
	stdout.Reset()
	if code := run(append(append([]string(nil), args[:2]...), args[3:]...), &stdout, &stderr); code != exitOK || !strings.Contains(stdout.String(), "partial: missing branch result [repair-and-replay]") {
		t.Fatal("entry values unavailable", code, stdout.String(), stderr.String())
	}
	if err := os.WriteFile(filepath.Join(copyRoot, "app.gooo.fixture"), []byte("package changed\n"), 0600); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	if code := run(args, &stdout, &stderr); code != exitFailure {
		t.Fatal("changed workspace replayed", code)
	}
}

func TestRunPackageReplayReportsUnmatchedCasesAsProgress(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	root := filepath.Join("..", "..", "examples", "package-imports")
	var stdout, stderr bytes.Buffer
	if code := run([]string{"package", "execute", "--json", "--cases", filepath.Join(root, "cases.json"), "--body-plans", filepath.Join(root, "body-plans.json"), filepath.Join(root, "gooo.workspace.json")}, &stdout, &stderr); code != exitOK {
		t.Fatal(code, stderr.String(), stdout.String())
	}
	path := filepath.Join(t.TempDir(), "saved.json")
	if err := os.WriteFile(path, stdout.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	if code := run([]string{"package", "replay", "--json", "--receipt", path, "--cases", filepath.Join(root, "cases.json"), filepath.Join(root, "gooo.workspace.json")}, &stdout, &stderr); code != exitOK {
		t.Fatal(code, stderr.String(), stdout.String())
	}
	var r packageExecutionReceipt
	if err := json.Unmarshal(stdout.Bytes(), &r); err != nil || r.Result == nil || r.Result.Replay == nil || r.Result.Replay.BodyFillsReplayed == 0 {
		t.Fatal("external package plans did not replay", err)
	}
	wrongCases := filepath.Join(t.TempDir(), "different-expectations.json")
	if err := os.WriteFile(wrongCases, []byte(`{"schema":"gooo/body-composition-cases/v1","cases":[{"inputs":{"example/core:Normalize":7},"expected":{"example/core:Normalize":9,"example/app:Main":9}}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	for _, operation := range []string{"replay", "execute"} {
		args := []string{"package", operation, "--json", "--cases", wrongCases, filepath.Join(root, "gooo.workspace.json")}
		if operation == "replay" {
			args = append(args, "--receipt", path)
		} else {
			args = append(args, "--body-plans", filepath.Join(root, "body-plans.json"))
		}
		stdout.Reset()
		if code := run(args, &stdout, &stderr); code != exitOK {
			t.Fatal(operation, code, stderr.String(), stdout.String())
		}
		if err := json.Unmarshal(stdout.Bytes(), &r); err != nil || r.Decision != "PROGRESS" || r.Result.Runtime.FinitePassed != 0 || r.Result.Runtime.FiniteTotal != 2 {
			t.Fatal("unmatched expectations became PASS", operation, err)
		}
	}
}

func TestRunPackageReplayRejectsAmbiguousOrInferenceOptions(t *testing.T) {
	for _, args := range [][]string{
		{"--receipt", "x", "--cases", "x", "--inputs", "x", "x"},
		{"--receipt", "x", "--cases", "x", "--assembly-model", "unused", "x"},
		{"--receipt", "x", "--cases", "x", "--receipt", "x", "x"},
		{"--cases", "x", "x"},
	} {
		var stdout, stderr bytes.Buffer
		if code := run(append([]string{"package", "replay"}, args...), &stdout, &stderr); code != exitUsage {
			t.Fatal("unsupported replay arguments accepted", args, code)
		}
	}
}
