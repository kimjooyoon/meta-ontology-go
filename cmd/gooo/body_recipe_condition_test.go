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

func TestConditionChainRecipeCLINativeExecution(t *testing.T) {
	source := "../../examples/body-codegen/condition-chain.gooo.fixture"
	plan := "../../examples/body-codegen/condition-chain.json"
	var generated, diagnostics bytes.Buffer
	if code := run([]string{"body-codegen", "--json", "--activity", "Clamp", "--path-plan", plan, source}, &generated, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String())
	}
	dir := t.TempDir()
	prior, cases := filepath.Join(dir, "generation.json"), filepath.Join(dir, "cases.json")
	if err := os.WriteFile(prior, generated.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	const suite = `{"schema":"gooo/body-runtime-cases/v1","cases":[{"input":-9223372036854775808,"expected":0},{"input":-1,"expected":0},{"input":0,"expected":0},{"input":5,"expected":5},{"input":10,"expected":10},{"input":11,"expected":10},{"input":9223372036854775807,"expected":10}]}`
	if err := os.WriteFile(cases, []byte(suite), 0600); err != nil {
		t.Fatal(err)
	}
	goBin := filepath.Join(runtime.GOROOT(), "bin", "go")
	if runtime.GOOS == "windows" {
		goBin += ".exe"
	}
	var executed bytes.Buffer
	if code := run([]string{"body-execute", "--source", source, "--path-plan", plan, "--generation", prior, "--cases", cases, "--go-bin", goBin}, &executed, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String())
	}
	var result bodyexecution.Result
	if err := json.Unmarshal(executed.Bytes(), &result); err != nil || result.Observation.Stage != "COMPLETE" || len(result.Observation.Cases) != 7 {
		t.Fatal("condition-chain runtime result", err)
	}
	for _, c := range result.Observation.Cases {
		if !c.Passed {
			t.Fatal("condition chain changed a range", c)
		}
	}
	if !strings.Contains(generated.String(), "canonical_typed_body_tree/v1") {
		t.Fatal("normalized chain omitted source-tree evidence")
	}
}
