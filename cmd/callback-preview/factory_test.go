package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/repositoryprojection/extractor"
)

func TestPreviewCLIFactoryUsesTypedLoweringInput(t *testing.T) {
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime caller unavailable")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(sourceFile), "../.."))
	output := filepath.Join(t.TempDir(), "factory.json")
	command := exec.Command("go", "run", "./cmd/callback-preview", "--root", root, "--strategy", "closure-factory", "--output", output)
	command.Dir = root
	if raw, err := command.CombinedOutput(); err != nil {
		t.Fatalf("callback factory CLI: %v\n%s", err, raw)
	}
	raw, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	var preview extractor.CallbackPreviewResult
	if err := json.Unmarshal(raw, &preview); err != nil {
		t.Fatal(err)
	}
	if err := extractor.ValidateCallbackPreviewResult(preview); err != nil {
		t.Fatal(err)
	}
	if preview.LoweringStrategy != "closure-factory" || preview.Candidate == nil || preview.ApplyPermission != "FORBIDDEN" {
		t.Fatalf("factory CLI output=%+v", preview)
	}
	bound := false
	for _, field := range preview.ContractRecords[0].Fields {
		if field.Name == "LoweringStrategy" && field.Value == preview.LoweringStrategy && field.ID == "gooo://meta-callback-preview/field/input-lowering-strategy" {
			bound = true
		}
	}
	if !bound {
		t.Fatal("CLI strategy was not recorded in the Gooo input field")
	}
}
