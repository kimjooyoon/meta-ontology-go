package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

func TestRunSourceProjectedRuntimePlanArtifactBoundary(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve typed runtime chain fixture path")
	}
	const sourcePath = "examples/language-runtime-binding/typed-chain.gooo"
	source, err := os.ReadFile(filepath.Join(filepath.Dir(currentFile), "..", "..", sourcePath))
	if err != nil {
		t.Fatal(err)
	}
	file, diagnostics := syntax.ParseFile(sourcePath, string(source))
	if diagnostics.HasErrors() || file == nil {
		t.Fatalf("typed chain parse diagnostics=%v file=%#v", diagnostics, file)
	}
	document, err := bidir.DocumentFromSyntax(file)
	if err != nil {
		t.Fatal(err)
	}
	typedPlan, err := bidir.CompileTypedPlan(document)
	if err != nil {
		t.Fatal(err)
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		t.Fatal(err)
	}
	generatedPlan, err := buildRuntimePlanDataWithTypedPlan(source, ir, typedPlan)
	if err != nil {
		t.Fatal(err)
	}
	var tamperedDocument runtimePlanDocument
	if err := json.Unmarshal(generatedPlan, &tamperedDocument); err != nil {
		t.Fatal(err)
	}
	tamperedDocument.TypedPlanDigest = "sha256:tampered"
	tamperedPlan, err := json.Marshal(tamperedDocument)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name        string
		runtimePlan []byte
		includePlan bool
		wantCode    string
		wantStage   string
		wantStep    string
	}{
		{
			name:        "missing generated artifact",
			includePlan: false,
			wantCode:    valueexecution.ReasonSourceReadFailed,
			wantStage:   "PLAN",
			wantStep:    "read-runtime-plan",
		},
		{
			name:        "tampered generated artifact",
			runtimePlan: tamperedPlan,
			includePlan: true,
			wantCode:    valueexecution.ReasonPlanInvalid,
			wantStage:   "PLAN",
			wantStep:    "validate-runtime-plan-contract",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			reader := runSourceReaderWithFiles{
				sourcePath:   source,
				"input.json": []byte(`{"value":41}`),
			}
			if testCase.includePlan {
				reader["runtime-plan.json"] = testCase.runtimePlan
			}
			var stdout, stderr bytes.Buffer
			code := runSource([]string{
				"--json", "--entry", "ProposeCandidate", "--input", "input.json",
				"--runtime-plan", "runtime-plan.json", sourcePath,
			}, reader, &stdout, &stderr)
			if code != exitFailure || stderr.Len() != 0 {
				t.Fatalf("code=%d stderr=%q stdout=%q", code, stderr.String(), stdout.String())
			}
			var report struct {
				Decision  string                   `json:"decision"`
				Reason    string                   `json:"reason"`
				Failure   valueexecution.Failure   `json:"failure"`
				Execution valueexecution.Execution `json:"execution"`
			}
			if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
				t.Fatal(err)
			}
			if report.Decision != "FAIL_CLOSED" || report.Reason != testCase.wantCode || report.Failure.Code != testCase.wantCode || report.Failure.Stage != testCase.wantStage || report.Failure.Step != testCase.wantStep || report.Execution.ApplyCalls != 0 || len(report.Execution.Activities) != 0 {
				t.Fatalf("projected runtime plan boundary report = %#v", report)
			}
		})
	}
}
