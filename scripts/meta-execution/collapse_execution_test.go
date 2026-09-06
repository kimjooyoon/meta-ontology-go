package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/linecaps"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestCollapseExecutionRouteUsesNativeMaterializer(t *testing.T) {
	root := collapseTestRepositoryRoot(t)
	head := collapseTestHead(t, root)
	subject := sourcepolicy.SourceSubject{Path: "scripts/meta-execution/collapsefixture/collapse_fixture.go", Line: 3, Name: "CollapseFixture"}
	plan, action, report := collapsePlannerFixture(t, subject.String(), head)
	if failure := validateCollapseAction(action); failure != nil {
		t.Fatalf("real planner action rejected: %#v", failure)
	}
	metricsPath := filepath.Join(t.TempDir(), "source-metrics.json")
	metricsPayload, err := json.Marshal(linecaps.LineMetricsReport{Root: root, CommitSHA: head, Meta: report})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(metricsPath, metricsPayload, 0o644); err != nil {
		t.Fatal(err)
	}
	fixtureWorkspace := filepath.Join(t.TempDir(), "workspace")
	if err := os.MkdirAll(fixtureWorkspace, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := copyTree(root, fixtureWorkspace); err != nil {
		t.Fatal(err)
	}
	if err := removeCollapseIntegrationTest(fixtureWorkspace); err != nil {
		t.Fatal(err)
	}
	gitDir, err := gitDirectory(root)
	if err != nil {
		t.Fatal(err)
	}
	materialized, failure := executeCollapse(fixtureWorkspace, gitDir, metricsPath, plan, action, metaExecutionTrace{})
	if failure != nil {
		t.Fatalf("native collapse execution failed: %#v", failure)
	}
	if materialized.OperationID != string(sourcepolicy.OperationCollapseAssign) || materialized.InstanceDigest == "" || materialized.ContractDigest == "" || materialized.Verifier.ExitCode != 0 {
		t.Fatalf("native collapse materialization = %#v", materialized)
	}
	var evidence collapseInstanceEvidence
	if err := json.Unmarshal(materialized.Canonical, &evidence); err != nil {
		t.Fatal(err)
	}
	if evidence.PreflightCount != 1 || evidence.ApplyCount != 1 || len(evidence.ChangedFiles) != 1 || evidence.ChangedFiles[0] != subject.Path || evidence.BeforePackageName != evidence.AfterPackageName || evidence.BeforeSignature != evidence.AfterSignature || len(evidence.BeforeCommentGroups) != len(evidence.AfterCommentGroups) {
		t.Fatalf("native collapse evidence = %#v", evidence)
	}
	if evidence.Process.Executor.ExitCode != 0 || evidence.Process.Evaluator.ExitCode != 0 || evidence.Process.Verifier.ExitCode != 0 {
		t.Fatalf("native collapse process evidence = %#v", evidence.Process)
	}
	beforeInspection, err := inspectCollapseSource(evidence.Before, subject)
	if err != nil {
		t.Fatal(err)
	}
	afterInspection, err := inspectCollapseSource(evidence.After, subject)
	if err != nil {
		t.Fatal(err)
	}
	if beforeInspection.StartLine != 3 || beforeInspection.EndLine != 6 || afterInspection.StartLine != 3 || afterInspection.EndLine >= beforeInspection.EndLine {
		t.Fatalf("native collapse span = before %d-%d after %d-%d", beforeInspection.StartLine, beforeInspection.EndLine, afterInspection.StartLine, afterInspection.EndLine)
	}
}

func collapseTestRepositoryRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("test source path unavailable")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
}

func collapseTestHead(t *testing.T, root string) string {
	t.Helper()
	output, err := exec.Command("git", "-C", root, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	head := strings.TrimSpace(string(output))
	if len(head) != 40 {
		t.Fatalf("invalid test head %q", head)
	}
	return head
}

func removeCollapseIntegrationTest(root string) error {
	return os.Remove(filepath.Join(root, filepath.FromSlash("scripts/meta-execution/collapse_execution_test.go")))
}
