package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspacesplit"
)

func TestMetricPlanReadsWorkspaceRequestIntoSelectedAction(t *testing.T) {
	root := t.TempDir()
	request := workspacesplit.Request{Workspace: "example/workspace.json", Package: "example", Source: "main.gooo",
		Into: "next.gooo", DiscoveryRoot: "example", Cases: "example/cases.json", LineLimit: 75}
	report, err := sourcepolicy.Evaluate(sourcepolicy.Default(), []sourcepolicy.Observation{
		{Subject: request.Subject(), Dimension: sourcepolicy.DimensionGoooFileLines, Value: 82, Producer: "linecaps.Analyze"},
		{Subject: "fixture.go:3:Collapse", Dimension: sourcepolicy.DimensionRefactorAssign, Value: 2, Producer: "linecaps.Analyze"},
	})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(map[string]any{"schema": "gooo/workspace-split-inputs/v1", "requests": []workspacesplit.Request{request}})
	if err != nil {
		t.Fatal(err)
	}
	inputs := filepath.Join(root, "requests.json")
	if err := os.WriteFile(inputs, raw, 0600); err != nil {
		t.Fatal(err)
	}
	opts := options{baseSHA: strings.Repeat("a", 40), sha: strings.Repeat("b", 40), plan: filepath.Join(root, "plan.json"), workspaceSplits: inputs}
	if err := emitGenerationPlan(opts, report); err != nil {
		t.Fatal(err)
	}
	raw, err = os.ReadFile(opts.plan)
	if err != nil {
		t.Fatal(err)
	}
	var plan generation.Plan
	if err := json.Unmarshal(raw, &plan); err != nil {
		t.Fatal(err)
	}
	for _, action := range plan.Selected {
		if action.Operation == sourcepolicy.OperationSplitGooo && action.WorkspaceSplitRequest != nil &&
			*action.WorkspaceSplitRequest == request && action.SourceIndicator.OperationInputDigest == request.Digest() {
			return
		}
	}
	t.Fatal("emitted plan lost its exact workspace request")
}
