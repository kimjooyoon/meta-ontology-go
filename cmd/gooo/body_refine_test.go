package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyrefinement"
)

func TestBodyRefinePublishesRoundProgramsAndRetainsOriginal(t *testing.T) {
	root := t.TempDir()
	source, err := os.ReadFile("../../examples/body-codegen/record-candidate-continuation.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	source, err = bodycodegen.ReviseAssemblyBudget(context.Background(), "target.gooo", source, "Select", 2)
	if err != nil {
		t.Fatal(err)
	}
	filename, output := filepath.Join(root, "target.gooo"), filepath.Join(root, "run")
	if err := os.WriteFile(filename, source, 0600); err != nil {
		t.Fatal(err)
	}
	args := []string{"body-refine", "--source", filename, "--activity", "Select", "--max-attempts", "2", "--max-rounds", "1",
		"--feedback-cases", "../../examples/body-codegen/record-field-updates-cases.json",
		"--policy", "../../examples/assembly-feedback/policy.gooo.fixture", "--out", output,
		"--go-bin", filepath.Join(runtime.GOROOT(), "bin", "go")}
	var stdout, stderr bytes.Buffer
	if code := run(args, &stdout, &stderr); code != exitOK {
		t.Fatal(code, stderr.String())
	}
	var result bodyrefinement.Result
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil || result.Status != "PROGRESS" || result.SelectedRound != 0 {
		t.Fatal("partial result was lost", err)
	}
	for _, relative := range []string{"refinement.json", "round-01/original.gooo", "selected/composition.json", "selected/realized.gooo", "selected/main.go"} {
		if _, err := os.Stat(filepath.Join(output, relative)); err != nil {
			t.Fatal(err)
		}
	}
	unchanged, _ := os.ReadFile(filename)
	if !bytes.Equal(unchanged, source) {
		t.Fatal("refinement overwrote the caller's original source")
	}
	stdout.Reset()
	if code := run(args, &stdout, &stderr); code != exitFailure {
		t.Fatal("existing output directory was overwritten")
	}
}
