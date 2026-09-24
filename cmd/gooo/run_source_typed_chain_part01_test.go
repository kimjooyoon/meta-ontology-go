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

func TestRunSourceTypedRuntimeChainHasIndependentReplayEvidence(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve typed runtime chain fixture path")
	}
	fixture := filepath.Join(filepath.Dir(currentFile), "..", "..", "examples", "language-runtime-binding", "typed-chain.gooo")
	source, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	const sourcePath = "examples/language-runtime-binding/typed-chain.gooo"
	run := func(runtimePlan []byte) typedChainReport {
		reader := runSourceReaderWithFiles{sourcePath: source, "input.json": []byte(`{"value":41}`)}
		args := []string{"--json", "--entry", "ProposeCandidate", "--input", "input.json"}
		if runtimePlan != nil {
			reader["runtime-plan.json"] = runtimePlan
			args = append(args, "--runtime-plan", "runtime-plan.json")
		}
		args = append(args, sourcePath)
		var stdout, stderr bytes.Buffer
		code := runSource(args, reader, &stdout, &stderr)
		if code != exitOK || stderr.Len() != 0 {
			t.Fatalf("code=%d stderr=%q stdout=%q", code, stderr.String(), stdout.String())
		}
		var report typedChainReport
		if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
			t.Fatal(err)
		}
		return report
	}
	first := run(nil)
	replay := run(nil)
	if first.Decision != "PASS" || first.Execution.Scope != valueexecution.RegisteredValueOperationScope || first.Execution.ApplyCalls != 3 || first.Execution.Deliveries != 2 || len(first.Execution.Activities) != 3 || first.Execution.Results["CommitCandidate"].Value != 44 {
		t.Fatalf("typed runtime chain report = %#v", first)
	}
	wantEdges := []string{
		"ProposeCandidate:result->RecordIndependentReview:input",
		"RecordIndependentReview:result->CommitCandidate:input",
	}
	if len(first.Execution.BindingEdgeOrder) != len(wantEdges) || first.Execution.BindingEdgeOrder[0] != wantEdges[0] || first.Execution.BindingEdgeOrder[1] != wantEdges[1] {
		t.Fatalf("typed runtime chain edge order = %#v, want %#v", first.Execution.BindingEdgeOrder, wantEdges)
	}
	firstExecution, err := json.Marshal(first.Execution)
	if err != nil {
		t.Fatal(err)
	}
	replayExecution, err := json.Marshal(replay.Execution)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstExecution, replayExecution) {
		t.Fatalf("independent replay changed execution evidence: first=%s replay=%s", firstExecution, replayExecution)
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
	runtimePlan, err := buildRuntimePlanDataWithTypedPlan(source, ir, typedPlan)
	if err != nil {
		t.Fatal(err)
	}
	planned := run(runtimePlan)
	if planned.RuntimePlanDigest == "" {
		t.Fatal("typed runtime plan digest was not retained in execution evidence")
	}
	if planned.Execution.RuntimePlanDigest != planned.RuntimePlanDigest {
		t.Fatalf("execution runtime plan digest = %q, want %q", planned.Execution.RuntimePlanDigest, planned.RuntimePlanDigest)
	}
	if len(planned.Execution.BindingEdgeOrder) != len(wantEdges) || planned.Execution.BindingEdgeOrder[0] != wantEdges[0] || planned.Execution.BindingEdgeOrder[1] != wantEdges[1] {
		t.Fatalf("planned runtime chain edge order = %#v, want %#v", planned.Execution.BindingEdgeOrder, wantEdges)
	}
}

type typedChainReport struct {
	Decision          string                   `json:"decision"`
	RuntimePlanDigest string                   `json:"runtime_plan_digest"`
	Execution         valueexecution.Execution `json:"execution"`
}
