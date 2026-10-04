package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

func TestSourceAssemblyCLIHasNoPlanFileAcrossGenerateExportAndExecute(t *testing.T) {
	source := "../../examples/body-codegen/source-assembly.gooo.fixture"
	var generated, diagnostics bytes.Buffer
	if code := run([]string{"body-codegen", "--json", "--activity", "Clamp", source}, &generated, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String())
	}
	var result bodycodegen.Result
	if err := json.Unmarshal(generated.Bytes(), &result); err != nil || result.Report.BodyPaths == nil ||
		result.Report.BodyPaths.FunctionalCompleteness != 100 || result.Report.BodyPaths.Search.Selection.ModelCalls != 0 {
		t.Fatal("source assembly did not use deterministic generation", err)
	}
	var exported bytes.Buffer
	if code := run([]string{"body-context", "--include-plan", "--activity", "Clamp", source}, &exported, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String())
	}
	var contextOutput bodyContextOutput
	if err := json.Unmarshal(exported.Bytes(), &contextOutput); err != nil || contextOutput.ExpandedPlan == nil ||
		len(contextOutput.Inputs) != 3 || contextOutput.ModelPredictions != 0 ||
		contextOutput.DocumentSHA256 != result.Report.BodyPaths.DocumentSHA256 {
		t.Fatal("source-only context export differs", err)
	}
	dir := t.TempDir()
	prior := filepath.Join(dir, "generation.json")
	if err := os.WriteFile(prior, generated.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	tool := filepath.Join(runtime.GOROOT(), "bin", "go")
	if runtime.GOOS == "windows" {
		tool += ".exe"
	}
	var executed bytes.Buffer
	if code := run([]string{"body-execute", "--source", source, "--generation", prior,
		"--cases", "../../examples/body-codegen/source-assembly-clamp-cases.json", "--go-bin", tool},
		&executed, &diagnostics); code != exitOK {
		t.Fatal(code, diagnostics.String())
	}
	var actual bodyexecution.Result
	if err := json.Unmarshal(executed.Bytes(), &actual); err != nil || actual.Observation.Stage != "COMPLETE" ||
		len(actual.Observation.Cases) != 7 || !actual.Observation.RuntimeReplayed {
		t.Fatal("source-only native execution failed", err)
	}
	for _, c := range actual.Observation.Cases {
		if !c.Passed {
			t.Fatal("clamp failed independent finite input", c)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	executed.Reset()
	if code := runBodyExecuteContext(ctx, []string{"--source", source, "--generation", prior,
		"--cases", "../../examples/body-codegen/source-assembly-clamp-cases.json"}, &executed, &diagnostics); code != exitFailure {
		t.Fatal("cancelled source-only execution continued")
	}
	if err := json.Unmarshal(executed.Bytes(), &actual); err != nil || actual.Observation.Toolchain.Started ||
		actual.Observation.ProjectionReplayed || actual.Observation.RuntimeReplayed {
		t.Fatal("cancelled assembly lost its unobserved execution record", err)
	}
}

func TestSourceAssemblyCLIPreservesIntentAuthorityAndMissingContractErrors(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/source-assembly.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	reader := mapSourceReader{"inline.gooo": source, "override.json": []byte(`{}`)}
	var stdout, stderr bytes.Buffer
	if code := runBodyCodegen([]string{"--json", "--activity", "Qualified", "--path-plan", "override.json", "inline.gooo"},
		reader, &stdout, &stderr); code != exitFailure || !strings.Contains(stdout.String(), "source assembling contract") {
		t.Fatal("external plan replaced source assembly", code, stdout.String())
	}
	for _, options := range [][]string{{"--sample-seed", "s"}, {"--fill-plan", "override.json"}, {"--fill-search", "override.json"}} {
		args := append([]string{"--activity", "Qualified", "inline.gooo"}, options...)
		stdout.Reset()
		stderr.Reset()
		if code := runBodyCodegen(args, reader, &stdout, &stderr); code != exitUsage {
			t.Fatal("conflicting generation modes accepted", args, code)
		}
	}
}
