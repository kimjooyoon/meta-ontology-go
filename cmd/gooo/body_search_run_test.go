package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"runtime"
	"testing"
)

func TestBodySearchRunGeneratesThenImmediatelyExecutesSourceSearch(t *testing.T) {
	t.Setenv("GOOO_LAYA_URL", "")
	goBin := filepath.Join(runtime.GOROOT(), "bin", "go")
	if runtime.GOOS == "windows" {
		goBin += ".exe"
	}
	var stdout, stderr bytes.Buffer
	args := []string{"--source", "../../examples/body-codegen/ir-search-source.gooo.fixture",
		"--activity", "ClampNegativeToZero", "--cases", "../../examples/body-codegen/ir-search-runtime-cases.json",
		"--go-bin", goBin}
	if code := runBodySearch(args, &stdout, &stderr); code != exitOK {
		t.Fatalf("body-search-run failed: code=%d stderr=%q", code, stderr.String())
	}
	var result bodySearchRunResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	search := result.Generation.Report.BodySearch
	if result.Schema != "gooo/body-search-run/v1" || search == nil || search.TrainingPassed != search.TrainingTotal ||
		search.HoldoutPassed != search.HoldoutTotal || search.CandidateGeneration == nil ||
		search.ProviderOperations != 0 {
		t.Fatalf("generation evidence was not retained: %#v", result.Generation.Report)
	}
	if result.Execution.Observation.Stage != "COMPLETE" || !result.Execution.Observation.RuntimeReplayed ||
		len(result.Execution.Observation.Cases) != 7 {
		t.Fatalf("runtime execution did not complete: %#v", result.Execution.Observation)
	}
	for _, c := range result.Execution.Observation.Cases {
		if !c.Passed {
			t.Fatalf("generated program failed independent runtime case: %#v", c)
		}
	}
	if result.Execution.CompletenessReceipt == nil {
		t.Fatal("runtime completeness receipt is missing")
	}
}

func TestBodySearchRunRejectsIncompleteInvocation(t *testing.T) {
	for _, args := range [][]string{nil, {"--source", "x"}, {"--unknown", "x"}} {
		var stdout, stderr bytes.Buffer
		if code := runBodySearch(args, &stdout, &stderr); code != exitUsage {
			t.Fatalf("accepted invalid invocation %q: code=%d", args, code)
		}
	}
}
