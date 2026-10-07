package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestBodyCompositionCLIPublishesSeparatedInputCases(t *testing.T) {
	tool := filepath.Join(runtime.GOROOT(), "bin", "go")
	if runtime.GOOS == "windows" {
		tool += ".exe"
	}
	directory := filepath.Join(t.TempDir(), "composition")
	var stdout, stderr bytes.Buffer
	args := []string{"body-compose", "--source", "../../examples/body-codegen/record-field-assembly.gooo.fixture",
		"--cases", "../../examples/body-codegen/record-field-assembly-cases.json", "--go-bin", tool, "--out", directory}
	if code := run(args, &stdout, &stderr); code != exitOK {
		t.Fatalf("compose(%d): %s", code, stderr.String())
	}
	var result bodyCompositionOutput
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	got := result.Runtime.InputSeparation
	if got.UniqueInputs != 7 || got.OverlappingInputs != 5 || got.DisjointInputs != 2 || got.DisjointCasesPassed != 2 || got.Status != "PASS" {
		t.Fatal("CLI input separation differs", got)
	}
	raw, err := os.ReadFile(filepath.Join(directory, "runtime.json"))
	if err != nil {
		t.Fatal(err)
	}
	var saved bodyexecution.CompositionRuntime
	if err := json.Unmarshal(raw, &saved); err != nil || saved.InputSeparation != got {
		t.Fatal("saved runtime lost input separation", err, saved.InputSeparation)
	}
}
