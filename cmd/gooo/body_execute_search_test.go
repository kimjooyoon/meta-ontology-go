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
)

func TestBodyExecuteSavedSourceSearchGeneration(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	source := "../../examples/body-codegen/ir-search-source.gooo.fixture"
	var generation, diagnostics bytes.Buffer
	if code := run([]string{"body-codegen", "--json", "--activity", "ClampNegativeToZero", source}, &generation, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String())
	}
	generationPath := filepath.Join(t.TempDir(), "generation.json")
	if err := os.WriteFile(generationPath, generation.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	goBinary := filepath.Join(runtime.GOROOT(), "bin", "go")
	if runtime.GOOS == "windows" {
		goBinary += ".exe"
	}
	var stdout, stderr bytes.Buffer
	if code := run([]string{"body-execute", "--source", source, "--generation", generationPath,
		"--cases", "../../examples/body-codegen/ir-search-runtime-cases.json", "--go-bin", goBinary}, &stdout, &stderr); code != exitOK {
		t.Fatalf("saved source search did not execute: code=%d stderr=%q", code, stderr.String())
	}
	var result bodyexecution.Result
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Observation.Stage != "COMPLETE" || !result.Observation.RuntimeReplayed || len(result.Observation.Cases) != 7 ||
		result.Observation.SelectionDisjointInputs != 0 {
		t.Fatalf("source-owned training/holdout and native evidence were not linked: %#v", result.Observation)
	}
	for _, observed := range result.Observation.Cases {
		if !observed.Passed {
			t.Fatalf("runtime case failed: %#v", observed)
		}
	}
	if _, err := bodyexecution.DecodeRuntimeReceipt(stdout.Bytes()); err != nil {
		t.Fatalf("native receipt could not be independently decoded: %v", err)
	}
	unchanged, err := os.ReadFile(generationPath)
	if err != nil || !bytes.Equal(unchanged, generation.Bytes()) {
		t.Fatal("saved generation was modified", err)
	}
	sourceBytes, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := bodyexecution.DecodeSourcePlan(context.Background(), source, sourceBytes, "ClampNegativeToZero", []byte(`{}`)); err == nil {
		t.Fatal("an external path document was silently ignored for source-owned IR search")
	}
	prior, _, err := bodyexecution.DecodeGeneration(generation.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	prior.Report.BodySearch.TrainingCaseResults[0].Input = 100
	changed, err := json.Marshal(prior)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(generationPath, changed, 0600); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	if code := run([]string{"body-execute", "--source", source, "--generation", generationPath,
		"--cases", "../../examples/body-codegen/ir-search-runtime-cases.json", "--go-bin", goBinary}, &stdout, &stderr); code != exitFailure {
		t.Fatal("changed finite evidence reached native execution")
	}
	result = bodyexecution.Result{}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil || result.Observation.Stage != "SOURCE_REPLAY" || result.Observation.Toolchain.Started || len(result.Observation.Runs) != 0 {
		t.Fatalf("replay failure did not stop before native execution: %v, %#v", err, result.Observation)
	}
}
