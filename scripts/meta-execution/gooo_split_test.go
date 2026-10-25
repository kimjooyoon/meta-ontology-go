package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspacesplit"
)

func goooSplitPlannerFixture(t *testing.T, head string) (generation.Plan, generation.Action) {
	t.Helper()
	request := workspacesplit.Request{Workspace: "examples/agent-tool-inventory/gooo.workspace.json",
		Package: "example/agent-tool-inventory", Source: "main.gooo", Into: "inspect.gooo",
		DiscoveryRoot: "examples/agent-tool-inventory", Cases: "examples/agent-tool-inventory/cases.json", LineLimit: 75}
	report, err := sourcepolicy.Evaluate(sourcepolicy.Default(), []sourcepolicy.Observation{
		{Subject: request.Subject(), Dimension: sourcepolicy.DimensionGoooFileLines, Value: 82, Producer: "linecaps.Analyze"},
		{Subject: "fixture.go:3:Collapse", Dimension: sourcepolicy.DimensionRefactorAssign, Value: 2, Producer: "linecaps.Analyze"},
	})
	if err != nil {
		t.Fatal(err)
	}
	report, inputs, err := generation.BindWorkspaceSplitInputs(report, []workspacesplit.Request{request})
	if err != nil {
		t.Fatal(err)
	}
	plan := generation.BuildWithOperationInputs(strings.Repeat("0", 40), head, report, inputs)
	if plan.Decision != generation.DecisionPlan {
		t.Fatal("missing independent metric plan", plan)
	}
	for _, action := range plan.Selected {
		if action.Operation == sourcepolicy.OperationSplitGooo {
			return plan, action
		}
	}
	t.Fatal("split was not selected")
	return plan, generation.Action{}
}

func TestGoooSplitExecutorUsesActualPublisherAndNativeOracles(t *testing.T) {
	repo, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command("git", "rev-parse", "HEAD")
	command.Dir = repo
	head, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	gitDir, err := gitDirectory(repo)
	if err != nil {
		t.Fatal(err)
	}
	plan, action := goooSplitPlannerFixture(t, strings.TrimSpace(string(head)))
	t.Setenv("GIT_WORK_TREE", repo)
	before, err := os.ReadFile(filepath.Join(repo, action.Subject))
	if err != nil {
		t.Fatal(err)
	}
	trace := metaExecutionTrace{state: newMetaExecutionTraceStateWithWriter(&bytes.Buffer{}), action: action}
	materialized, failure := executeGoooSplit(repo, gitDir, plan, action, trace)
	if failure != nil {
		t.Fatalf("native split failed: %v diagnostics=%v", failure, failure.diagnostics)
	}
	var instance goooSplitReplay
	if err := json.Unmarshal(materialized.Canonical, &instance); err != nil {
		t.Fatal(err)
	}
	if instance.Runtime.FinitePassed != 8 || instance.Runtime.FiniteTotal != 8 || len(instance.Runtime.Traces) != 8 ||
		len(materialized.Indicators) != 5 || materialized.InstanceDigest == "" || materialized.Verifier.ExitCode != 0 {
		t.Fatal("native split did not preserve the finite oracle and complete execution evidence")
	}
	after, err := os.ReadFile(filepath.Join(repo, action.Subject))
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("disposable operation changed caller source", err)
	}
	if _, err := os.Stat(filepath.Join(repo, "examples/agent-tool-inventory/inspect.gooo")); !os.IsNotExist(err) {
		t.Fatal("caller received split output", err)
	}
	t.Logf("source %s; finite cases %d/%d; two fresh Git-tree passes; selected operation %s", plan.HeadSHA, instance.Runtime.FinitePassed, instance.Runtime.FiniteTotal, materialized.OperationID)
}

func TestGoooSplitExecutorMissingScopeIsUnresolved(t *testing.T) {
	_, action := goooSplitPlannerFixture(t, strings.Repeat("a", 40))
	action.WorkspaceSplitRequest = nil
	_, failure := executeGoooSplit("", "", generation.Plan{}, action, metaExecutionTrace{})
	if failure == nil || failure.reason != "WORKSPACE_SPLIT_REQUEST_MISSING" || failure.class != "DIRECT_MISSING" {
		t.Fatal("missing consumer scope was not retained", failure)
	}
}
