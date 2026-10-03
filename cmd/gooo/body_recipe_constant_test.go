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

func TestConstantRecipeCLINativeExecution(t *testing.T) {
	source := "../../examples/body-codegen/constant-recipe.gooo.fixture"
	plan := "../../examples/body-codegen/constant-recipe.json"
	var generated, diagnostics bytes.Buffer
	if code := run([]string{"body-codegen", "--json", "--activity", "Constant", "--path-plan", plan, source}, &generated, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String())
	}
	dir := t.TempDir()
	prior, cases := filepath.Join(dir, "generation.json"), filepath.Join(dir, "cases.json")
	if err := os.WriteFile(prior, generated.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cases, []byte(`{"schema":"gooo/body-runtime-cases/v1","cases":[{"input":-9223372036854775808,"expected":1},{"input":0,"expected":1},{"input":9223372036854775807,"expected":1}]}`), 0600); err != nil {
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
	if err := json.Unmarshal(executed.Bytes(), &result); err != nil || result.Observation.Stage != "COMPLETE" || len(result.Observation.Cases) != 3 {
		t.Fatal("constant recipe runtime result", err)
	}
	for _, c := range result.Observation.Cases {
		if !c.Passed {
			t.Fatal("compiled constant depends on its unread input", c)
		}
	}
}
