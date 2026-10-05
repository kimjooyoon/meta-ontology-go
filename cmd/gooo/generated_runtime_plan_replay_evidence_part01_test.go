package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/lsp"
	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

func TestGeneratedRuntimePlanReplaysFromCleanProjectedWorkspaces(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve typed runtime chain fixture path")
	}
	const sourcePath = "examples/language-runtime-binding/typed-chain.gooo"
	source, err := os.ReadFile(filepath.Join(filepath.Dir(currentFile), "..", "..", sourcePath))
	if err != nil {
		t.Fatal(err)
	}

	generatedRoot := t.TempDir()
	var generateStdout, generateStderr bytes.Buffer
	generateCode := runGenerate([]string{
		sourcePath,
		"--out", generatedRoot,
	}, runSourceReaderWithFiles{sourcePath: source}, EntityFieldsCLIParser{}, &generateStdout, &generateStderr)
	if generateCode != exitOK || generateStderr.Len() != 0 {
		t.Fatalf("generate code=%d stderr=%q stdout=%q", generateCode, generateStderr.String(), generateStdout.String())
	}
	runtimePlan, err := os.ReadFile(filepath.Join(generatedRoot, "runtime-plan.json"))
	if err != nil {
		t.Fatal(err)
	}
	generatedGo, err := os.ReadFile(filepath.Join(generatedRoot, generatedFileName))
	if err != nil || len(generatedGo) == 0 {
		t.Fatalf("generated executable artifact missing: bytes=%d err=%v", len(generatedGo), err)
	}
	var planDocument runtimePlanDocument
	if err := json.Unmarshal(runtimePlan, &planDocument); err != nil {
		t.Fatal(err)
	}

	type replayReport struct {
		Decision                string                                       `json:"decision"`
		RuntimePlanDigest       string                                       `json:"runtime_plan_digest"`
		GeneratedReplayEvidence valueexecution.GeneratedReplayEvidencePart01 `json:"generated_replay_evidence"`
		ExecutionOriginReceipt  valueexecution.ExecutionOriginReceipt        `json:"execution_origin_receipt"`
		Execution               valueexecution.Execution                     `json:"execution"`
	}
	runProjected := func(plan []byte) replayReport {
		t.Helper()
		workspace := t.TempDir()
		sourceFilename := filepath.Join(workspace, filepath.FromSlash(sourcePath))
		if err := os.MkdirAll(filepath.Dir(sourceFilename), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(sourceFilename, source, 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(workspace, "input.json"), []byte(`{"value":41}`), 0o600); err != nil {
			t.Fatal(err)
		}
		if plan != nil {
			if err := os.WriteFile(filepath.Join(workspace, "runtime-plan.json"), plan, 0o600); err != nil {
				t.Fatal(err)
			}
		}
		var stdout, stderr bytes.Buffer
		code := runSource([]string{
			"--json", "--entry", "ProposeCandidate", "--input", "input.json",
			"--runtime-plan", "runtime-plan.json", sourcePath,
		}, rootedRunSourceReader{root: workspace}, &stdout, &stderr)
		if code != exitOK || stderr.Len() != 0 {
			t.Fatalf("projected replay code=%d stderr=%q stdout=%q", code, stderr.String(), stdout.String())
		}
		var report replayReport
		if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
			t.Fatal(err)
		}
		return report
	}

	first := runProjected(runtimePlan)
	second := runProjected(runtimePlan)
	if first.Decision != "PASS" || first.Execution.ApplyCalls != 3 || len(first.Execution.Activities) != 3 ||
		first.Execution.Results["CommitCandidate"].Value != 44 {
		t.Fatalf("projected execution=%#v", first)
	}
	if !reflect.DeepEqual(first.Execution, second.Execution) ||
		first.GeneratedReplayEvidence != second.GeneratedReplayEvidence ||
		!reflect.DeepEqual(first.ExecutionOriginReceipt, second.ExecutionOriginReceipt) {
		t.Fatalf("independent replay changed its sealed evidence: first=%#v second=%#v", first, second)
	}

	evidence := first.GeneratedReplayEvidence
	wantArtifactDigest := "sha256:" + cache.HashBytes(runtimePlan).String()
	if evidence.SourceDigest != planDocument.SourceDigest || evidence.SemanticDigest != planDocument.SemanticHash ||
		evidence.TypedPlanDigest != planDocument.TypedPlanDigest || evidence.RuntimePlanDigest != first.RuntimePlanDigest ||
		evidence.GeneratedArtifactDigest != wantArtifactDigest || evidence.ToolchainDigest == "" ||
		evidence.EvaluatorDigest == "" || evidence.ReverseObservationDigest != first.ExecutionOriginReceipt.ReceiptDigest ||
		first.ExecutionOriginReceipt.ExecutionDigest != first.Execution.ExecutionDigest ||
		first.ExecutionOriginReceipt.RuntimePlanDigest != first.RuntimePlanDigest || !first.ExecutionOriginReceipt.Validate() {
		t.Fatalf("generated replay identities are incomplete or inconsistent: evidence=%#v receipt=%#v execution=%#v", evidence, first.ExecutionOriginReceipt, first.Execution)
	}
	closure := lsp.ObserveGeneratedReplayEvidenceReceiptClosurePart01(evidence, first.ExecutionOriginReceipt)
	if closure.Status != lsp.ExecutionEvidenceReceiptClosureComplete || closure.MissingStageIndex != -1 ||
		!lsp.ValidateGeneratedReplayEvidenceReceiptClosurePart01(closure, evidence, first.ExecutionOriginReceipt) {
		t.Fatalf("current generated replay did not close existing LSP evidence: %+v", closure)
	}

	continuationWorkspace := t.TempDir()
	continuationSource := filepath.Join(continuationWorkspace, filepath.FromSlash(sourcePath))
	if err := os.MkdirAll(filepath.Dir(continuationSource), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(continuationSource, source, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(continuationWorkspace, "input.json"), []byte(`{"value":41}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(continuationWorkspace, "runtime-plan.json"), runtimePlan, 0o600); err != nil {
		t.Fatal(err)
	}
	var continuationStdout, continuationStderr bytes.Buffer
	continuationCode := runSource([]string{
		"--json", "--entry", "ProposeCandidate", "--input", "input.json",
		"--runtime-plan", "runtime-plan.json", "--iterations", "1", sourcePath,
	}, rootedRunSourceReader{root: continuationWorkspace}, &continuationStdout, &continuationStderr)
	if continuationCode != exitOK || continuationStderr.Len() != 0 {
		t.Fatalf("continuation replay code=%d stderr=%q stdout=%q", continuationCode, continuationStderr.String(), continuationStdout.String())
	}
	var continuationReport struct {
		Decision                string                                       `json:"decision"`
		GeneratedReplayEvidence valueexecution.GeneratedReplayEvidencePart01 `json:"generated_replay_evidence"`
		ExecutionOriginReceipt  valueexecution.ExecutionOriginReceipt        `json:"execution_origin_receipt"`
		Continuation            valueexecution.Continuation                  `json:"continuation"`
	}
	if err := json.Unmarshal(continuationStdout.Bytes(), &continuationReport); err != nil {
		t.Fatal(err)
	}
	continuationClosure := lsp.ObserveGeneratedReplayEvidenceReceiptClosurePart01(
		continuationReport.GeneratedReplayEvidence,
		continuationReport.ExecutionOriginReceipt,
	)
	if continuationReport.Decision != "PASS" || len(continuationReport.Continuation.Executions) != 1 ||
		continuationReport.GeneratedReplayEvidence.ReverseObservationDigest != continuationReport.ExecutionOriginReceipt.ReceiptDigest ||
		continuationClosure.Status != lsp.ExecutionEvidenceReceiptClosureComplete {
		t.Fatalf("generated replay evidence missing from continuation output: report=%#v closure=%+v", continuationReport, continuationClosure)
	}

	tamperedDocument := planDocument
	tamperedDocument.TypedPlanDigest = "sha256:" + cache.HashBytes([]byte("tampered typed plan")).String()
	tamperedPlan, err := json.Marshal(tamperedDocument)
	if err != nil {
		t.Fatal(err)
	}
	for _, testCase := range []struct {
		name     string
		plan     []byte
		wantStep string
		wantCode string
	}{
		{name: "missing plan artifact", plan: nil, wantStep: "read-runtime-plan", wantCode: valueexecution.ReasonSourceReadFailed},
		{name: "tampered plan artifact", plan: tamperedPlan, wantStep: "validate-runtime-plan-contract", wantCode: valueexecution.ReasonPlanInvalid},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			workspace := t.TempDir()
			sourceFilename := filepath.Join(workspace, filepath.FromSlash(sourcePath))
			if err := os.MkdirAll(filepath.Dir(sourceFilename), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(sourceFilename, source, 0o600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(workspace, "input.json"), []byte(`{"value":41}`), 0o600); err != nil {
				t.Fatal(err)
			}
			if testCase.plan != nil {
				if err := os.WriteFile(filepath.Join(workspace, "runtime-plan.json"), testCase.plan, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			var stdout, stderr bytes.Buffer
			code := runSource([]string{
				"--json", "--entry", "ProposeCandidate", "--input", "input.json",
				"--runtime-plan", "runtime-plan.json", sourcePath,
			}, rootedRunSourceReader{root: workspace}, &stdout, &stderr)
			if code != exitFailure || stderr.Len() != 0 {
				t.Fatalf("rejected replay code=%d stderr=%q stdout=%q", code, stderr.String(), stdout.String())
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
			if report.Decision != "FAIL_CLOSED" || report.Reason != testCase.wantCode ||
				report.Failure.Step != testCase.wantStep || report.Execution.ApplyCalls != 0 || len(report.Execution.Activities) != 0 {
				t.Fatalf("rejected artifact crossed execution frontier: %#v", report)
			}
		})
	}
}

type rootedRunSourceReader struct{ root string }

func (reader rootedRunSourceReader) ReadFile(name string) ([]byte, error) {
	if filepath.IsAbs(name) {
		return nil, os.ErrPermission
	}
	cleaned := filepath.Clean(name)
	if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return nil, os.ErrPermission
	}
	return os.ReadFile(filepath.Join(reader.root, cleaned))
}
