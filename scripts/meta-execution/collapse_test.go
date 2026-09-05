package main

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/linecaps"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestCollapseDispatchKeepsUnsupportedOperationFailClosed(t *testing.T) {
	_, failure := executeAction("", "", "", generation.Plan{}, generation.Action{Operation: sourcepolicy.OperationObserve}, metaExecutionTrace{})
	if failure == nil || failure.reason != "UNSUPPORTED_SELECTED_OPERATION" {
		t.Fatalf("unsupported operation failure = %#v, want fail-closed unsupported selection", failure)
	}
}


func TestCollapseDispatchRejectsSingleFieldBindingTampering(t *testing.T) {
	action := collapsePlannerAction(t, "fixture.go:3:value", strings.Repeat("1", 40))
	if failure := validateCollapseAction(action); failure != nil {
		t.Fatalf("real planner action rejected: %#v", failure)
	}
	mutations := []struct {
		name   string
		mutate func(*generation.Action)
	}{
		{name: "executor", mutate: func(candidate *generation.Action) { candidate.Executor = "tampered-executor" }},
		{name: "evaluator", mutate: func(candidate *generation.Action) { candidate.Evaluator = "tampered-evaluator" }},
		{name: "source contract digest", mutate: func(candidate *generation.Action) { candidate.InputContractSourceDigest = strings.Repeat("0", 64) }},
		{name: "semantic contract digest", mutate: func(candidate *generation.Action) { candidate.InputContractSemanticDigest = strings.Repeat("0", 64) }},
		{name: "required indicator", mutate: func(candidate *generation.Action) {
			candidate.RequiredIndicatorIDs = append([]string{}, candidate.RequiredIndicatorIDs...)
			candidate.RequiredIndicatorIDs[0] = "tampered-indicator"
		}},
		{name: "metric producer", mutate: func(candidate *generation.Action) { candidate.MetricProducer = "tampered-producer" }},
		{name: "source proof", mutate: func(candidate *generation.Action) { candidate.SourceIndicator.Proof = sourcepolicy.ProofCoherence }},
		{name: "indicator outcome", mutate: func(candidate *generation.Action) { candidate.IndicatorOutcome.Decision = sourcepolicy.IndicatorDecisionPass }},
	}
	for _, mutation := range mutations {
		t.Run(mutation.name, func(t *testing.T) {
			candidate := action
			mutation.mutate(&candidate)
			if failure := validateCollapseAction(candidate); failure == nil || failure.reason != "ACTION_BINDING_INVALID" {
				t.Fatalf("tampered %s validation = %#v, want fail-closed binding rejection", mutation.name, failure)
			}
		})
	}
}

func TestInspectCollapseSourceBindsExactSubjectAndReceiver(t *testing.T) {
	source := []byte("package fixture\n\ntype receiver struct{}\n\nfunc (receiver) value() int {\n\tresult := 1\n\treturn result\n}\n")
	subject := sourcepolicy.SourceSubject{Path: "fixture.go", Line: 5, Name: "method value"}
	before, err := inspectCollapseSource(source, subject)
	if err != nil {
		t.Fatal(err)
	}
	if !before.AssignmentReturn || !before.CommentsPreserved || before.Receiver != "(receiver)" {
		t.Fatalf("before inspection = %#v", before)
	}
	afterSource := []byte("package fixture\n\ntype receiver struct{}\n\nfunc (receiver) value() int {\n\treturn 1\n}\n")
	after, err := inspectCollapseSource(afterSource, subject)
	if err != nil {
		t.Fatal(err)
	}
	if !after.SingleReturn || after.ReturnExpression != before.ReturnExpression || after.Receiver != before.Receiver {
		t.Fatalf("after inspection = %#v", after)
	}
	if failure := validateCollapseOutput(before, after, afterSource); failure != nil {
		t.Fatalf("valid collapse rejected: %#v", failure)
	}
}

func TestInspectCollapseSourceRejectsBodylessDeclaration(t *testing.T) {
	source := []byte("package fixture\n\nfunc value() int\n")
	if _, err := inspectCollapseSource(source, sourcepolicy.SourceSubject{Path: "fixture.go", Line: 3, Name: "value"}); err == nil {
		t.Fatal("bodyless declaration was accepted")
	}
}

func TestCollapseExecutionRouteUsesNativeMaterializer(t *testing.T) {
	root := collapseTestRepositoryRoot(t)
	head := collapseTestHead(t, root)
	subject := collapseTestSubject(t, root)
	plan, action, report := collapsePlannerFixture(t, subject, head)
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
	if err := removeGoTests(fixtureWorkspace); err != nil {
		t.Fatal(err)
	}
	gitDir, err := gitDirectory(root)
	if err != nil {
		t.Fatal(err)
	}
	materialized, failure := executeCollapse(fixtureWorkspace, gitDir, metricsPath, plan, action, subject, metaExecutionTrace{}, "positive")
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
	if evidence.PreflightCount != 1 || evidence.ApplyCount != 1 || len(evidence.ChangedFiles) != 1 || evidence.ChangedFiles[0] != subject.Path || evidence.BeforeSignature != evidence.AfterSignature || len(evidence.BeforeCommentGroups) != len(evidence.AfterCommentGroups) {
		t.Fatalf("native collapse evidence = %#v", evidence)
	}
	if evidence.Process.Executor.ExitCode != 0 || evidence.Process.Evaluator.ExitCode != 0 || evidence.Process.Verifier.ExitCode != 0 {
		t.Fatalf("native collapse process evidence = %#v", evidence.Process)
	}
}

func collapsePlannerAction(t *testing.T, subject, head string) generation.Action {
	t.Helper()
	_, action, _ := collapsePlannerFixture(t, subject, head)
	return action
}

func collapsePlannerFixture(t *testing.T, subject, head string) (generation.Plan, generation.Action, sourcepolicy.Report) {
	t.Helper()
	report, err := sourcepolicy.Evaluate(sourcepolicy.Default(), []sourcepolicy.Observation{
		{Subject: subject, Dimension: sourcepolicy.DimensionRefactorAssign, Value: 2, Detail: "assignment then return result", Producer: "linecaps.Analyze"},
		{Subject: "operations.go:1:executeAction", Dimension: sourcepolicy.DimensionFunctionLines, Value: 76, Producer: "linecaps.Analyze"},
	})
	if err != nil {
		t.Fatal(err)
	}
	plan := generation.Build(strings.Repeat("0", 40), head, report)
	for _, action := range plan.Selected {
		if action.Operation == sourcepolicy.OperationCollapseAssign {
			return plan, action, report
		}
	}
	t.Fatalf("planner did not select collapse action: %#v", plan)
	return generation.Plan{}, generation.Action{}, sourcepolicy.Report{}
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

func collapseTestSubject(t *testing.T, root string) string {
	t.Helper()
	logicalPath := "scripts/meta-execution/collapse.go"
	path := filepath.Join(root, filepath.FromSlash(logicalPath))
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, logicalPath, source, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	var subject sourcepolicy.SourceSubject
	ast.Inspect(file, func(node ast.Node) bool {
		function, ok := node.(*ast.FuncDecl)
		if !ok || subject.Path != "" {
			return true
		}
		name, line, identityOK := linecaps.FunctionIdentity(fset, function)
		if identityOK && name == "collapseSummaryMatches" {
			subject = sourcepolicy.SourceSubject{Path: logicalPath, Line: line, Name: name}
		}
		return true
	})
	if subject.Path == "" {
		t.Fatal("collapse test candidate function not found")
	}
	return subject.String()
}

func removeGoTests(root string) error {
	return filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if strings.HasSuffix(entry.Name(), "_test.go") {
			return os.Remove(path)
		}
		return nil
	})
}
