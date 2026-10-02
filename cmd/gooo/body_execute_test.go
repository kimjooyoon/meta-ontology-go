package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/completeness"
	"github.com/kimjooyoon/meta-ontology-go/internal/completenessdelta"
)

func TestBodyExecuteCLIGenerationToNativeObservation(t *testing.T) {
	dir := t.TempDir()
	source := "../../examples/body-codegen/typed-path-compound.gooo.fixture"
	plan := "../../examples/body-codegen/typed-path-compound-plan.json"
	var generated, diagnostics bytes.Buffer
	if code := run([]string{"body-codegen", "--json", "--activity", "Combined", "--path-plan", plan, source}, &generated, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String())
	}
	priorPath, casesPath := filepath.Join(dir, "generation.json"), filepath.Join(dir, "cases.json")
	if err := os.WriteFile(priorPath, generated.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(casesPath, []byte(`{"schema":"gooo/body-runtime-cases/v1","cases":[{"input":4,"expected":25},{"input":-4,"expected":999}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	tool := filepath.Join(runtime.GOROOT(), "bin", "go")
	if runtime.GOOS == "windows" {
		tool += ".exe"
	}
	args := []string{"body-execute", "--source", source, "--path-plan", plan, "--generation", priorPath, "--cases", casesPath, "--go-bin", tool}
	var stdout, stderr bytes.Buffer
	if code := run(args, &stdout, &stderr); code != exitOK {
		t.Fatal(code, stderr.String())
	}
	var result bodyexecution.Result
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Observation.Stage != "COMPLETE" || len(result.Observation.Cases) != 2 || result.Observation.Cases[1].Passed {
		t.Fatal("CLI lost partial native observation")
	}
	raw, _ := json.Marshal(result.CompletenessReceipt)
	if _, err := completeness.Decode(raw); err != nil {
		t.Fatal(err)
	}
	runtimePath := filepath.Join(dir, "runtime.json")
	if err := os.WriteFile(runtimePath, stdout.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	var deltaOut, deltaErr bytes.Buffer
	if code := run([]string{"completeness-delta", "--before", priorPath, "--after", runtimePath}, &deltaOut, &deltaErr); code != exitOK {
		t.Fatal(code, deltaErr.String())
	}
	var delta completenessdelta.CompletenessDelta
	if err := json.Unmarshal(deltaOut.Bytes(), &delta); err != nil || delta.Relation != "PARENT_RUNTIME_CONTINUATION" {
		t.Fatal("generation/runtime linkage lost", err, delta.Relation)
	}
	if delta.ComparatorOperations["model_calls"] != 0 {
		t.Fatal("comparison called model")
	}
	for _, d := range delta.Dimensions {
		if d.CountMagnitude != nil {
			t.Fatal("cross-profile numeric delta")
		}
	}
	after, err := os.ReadFile(priorPath)
	if err != nil || !bytes.Equal(after, generated.Bytes()) {
		t.Fatal("CLI rewrote the original generation")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	stdout.Reset()
	stderr.Reset()
	if code := runBodyExecuteContext(ctx, args[1:], &stdout, &stderr); code != exitFailure {
		t.Fatal("cancellation ignored")
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil || result.Observation.Toolchain.Started {
		t.Fatal("canceled CLI started a tool", err)
	}
}

func TestBodyExecuteCLIInputBounds(t *testing.T) {
	for _, args := range [][]string{nil, {"--unknown", "x"}, {"--source"}, {"--source", "a", "--source", "b"}} {
		var out, err bytes.Buffer
		if code := runBodyExecute(args, &out, &err); code != exitUsage {
			t.Fatal("accepted invalid arguments", args, code)
		}
	}
	if _, err := readBodyExecutionFile(t.TempDir(), 10); err == nil {
		t.Fatal("directory accepted")
	}
	path := filepath.Join(t.TempDir(), "too-large")
	if err := os.WriteFile(path, []byte("123456"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := readBodyExecutionFile(path, 5); err == nil {
		t.Fatal("oversize accepted")
	}
}
