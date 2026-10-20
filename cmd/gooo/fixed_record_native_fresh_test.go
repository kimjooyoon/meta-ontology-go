package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/projectionbundle"
)

func TestFixedRecordNativeFresh1767(t *testing.T) {
	if os.Getenv("GOOO_FIXED_RECORD_NATIVE_E2E") != "1" {
		t.Skip("original CI owns this native source-to-runtime observation")
	}
	t.Setenv("GOOO_LAYA_URL", "")
	t.Setenv("GOOO_LAYA_API_KEY", "")
	directory := os.Getenv("GOOO_FIXED_RECORD_NATIVE_OUTPUT")
	if directory == "" {
		t.Fatal("native observation requires a new caller-owned output directory")
	}
	if err := os.MkdirAll(filepath.Dir(directory), 0750); err != nil {
		t.Fatal(err)
	}
	source := "../../examples/fixed-record-contract/main.gooo"
	raw, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	tool := filepath.Join(runtime.GOROOT(), "bin", "go")
	if runtime.GOOS == "windows" {
		tool += ".exe"
	}
	var stdout, stderr bytes.Buffer
	args := []string{"body-compose", "--source", source, "--entry", "Decide", "--cases",
		"../../examples/fixed-record-contract/evaluation.json", "--go-bin", tool, "--out", directory}
	if code := run(args, &stdout, &stderr); code != exitOK {
		t.Fatalf("fixed source native command(%d): %s %s", code, &stderr, &stdout)
	}
	var output bodyCompositionOutput
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatal(err)
	}
	r := output.Runtime
	expectedGo := "go version go1.27.3 " + runtime.GOOS + "/" + runtime.GOARCH
	if !output.GeneratedNow || output.Composition.Stage != "COMPLETE" || r.Stage != "COMPLETE" ||
		r.FinitePassed != 4 || r.FiniteTotal != 4 || len(r.Runs) != 2 || !r.RuntimeReplayed || !r.ProjectionReplayed ||
		r.ModelCalls != 0 || r.OriginalSourceSHA256 != projectionbundle.Digest(raw) || r.GoVersion != expectedGo ||
		len(output.Composition.Plan.Activities) != 1 || len(output.Composition.Plan.Edges) != 0 ||
		len(output.Composition.Steps) != 1 {
		t.Fatalf("native source scope or output comparison changed: %+v", r)
	}
	contract := output.Composition.Steps[0].Generation.Report.RecordAssembly
	if contract == nil || contract.Total != 4 || contract.Passed != 4 || contract.FieldsTotal != 12 ||
		contract.FieldsPassed != 12 || len(contract.Attempts) != 1 || len(contract.Choices) != 0 || contract.ModelCalls != 0 {
		t.Fatalf("native composition lost its fixed source contract: %+v", contract)
	}
	for _, process := range r.Runs {
		if !process.Completed || process.Canceled || process.TimedOut || process.ExitCode == nil || *process.ExitCode != 0 {
			t.Fatalf("native child did not complete normally: %+v", process)
		}
	}
	if err := os.WriteFile(filepath.Join(directory, "command.json"), stdout.Bytes(), 0640); err != nil {
		t.Fatal(err)
	}
}
