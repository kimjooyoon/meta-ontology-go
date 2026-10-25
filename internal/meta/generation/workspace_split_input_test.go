package generation

import (
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspacesplit"
)

func workspaceSplitPlanFixture() (sourcepolicy.Report, workspacesplit.Request) {
	request := workspacesplit.Request{Workspace: "examples/sample/workspace.json", Package: "sample",
		Source: "main.gooo", Into: "part.gooo", DiscoveryRoot: "examples/sample", Cases: "examples/sample/cases.json", LineLimit: 75}
	return sourcepolicy.Report{Schema: sourcepolicy.IndicatorSchema, Policy: sourcepolicy.Default(), Indicators: []sourcepolicy.Indicator{
		metric(request.Subject(), sourcepolicy.OperationSplitGooo, false, false),
		metric("fixture.go:3:Collapse", sourcepolicy.OperationCollapseAssign, false, false)}}, request
}

func TestWorkspaceSplitPlanPinsScopeAndKeepsIndependentFloor(t *testing.T) {
	original, request := workspaceSplitPlanFixture()
	report, inputs, err := BindWorkspaceSplitInputs(original, []workspacesplit.Request{request})
	if err != nil {
		t.Fatal(err)
	}
	base, head := strings.Repeat("a", 40), strings.Repeat("b", 40)
	plan := BuildWithOperationInputs(base, head, report, inputs)
	if plan.Decision != DecisionPlan || len(plan.Selected) != 2 || original.Indicators[0].OperationInputDigest != "" {
		t.Fatal("typed scope mutated the observation or changed the independent floor", plan)
	}
	manifest := BuildExecutionManifest(plan)
	if manifest.Decision != ExecutionDecisionProposed {
		t.Fatal(manifest)
	}
	count := 0
	for _, step := range manifest.Steps {
		if step.Operation != sourcepolicy.OperationSplitGooo {
			continue
		}
		count++
		if !reflect.DeepEqual(step.WorkspaceSplitRequest, &request) || step.SourceIndicator.OperationInputDigest != request.Digest() {
			t.Fatal("execution lost its exact scoped request", step)
		}
	}
	if count != 1 {
		t.Fatal("missing split step")
	}
	request.DiscoveryRoot = "examples"
	changed, changedInputs, err := BindWorkspaceSplitInputs(original, []workspacesplit.Request{request})
	if err != nil {
		t.Fatal(err)
	}
	other := BuildWithOperationInputs(base, head, changed, changedInputs)
	if other.InputDigest == plan.InputDigest || other.PlanDigest == plan.PlanDigest {
		t.Fatal("scope change reused plan identity")
	}
	if replay := BuildWithOperationInputs(base, head, report, inputs); !reflect.DeepEqual(plan, replay) {
		t.Fatal("scope did not replay")
	}
}

func TestWorkspaceSplitPlanRetainsUnboundAndContradictoryInputs(t *testing.T) {
	original, request := workspaceSplitPlanFixture()
	report, inputs, err := BindWorkspaceSplitInputs(original, []workspacesplit.Request{request})
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"missing", "stale", "wrong-limit", "unreferenced"} {
		local := OperationInputs{WorkspaceSplits: map[string]workspacesplit.Request{}}
		for k, v := range inputs.WorkspaceSplits {
			local.WorkspaceSplits[k] = v
		}
		modified := request
		switch mode {
		case "missing":
			delete(local.WorkspaceSplits, request.Digest())
		case "stale":
			modified.Cases = "other.json"
			local.WorkspaceSplits[request.Digest()] = modified
		case "wrong-limit":
			modified.LineLimit++
			bad := original
			bad.Indicators = append([]sourcepolicy.Indicator(nil), original.Indicators...)
			bad.Indicators[0].OperationInputDigest = modified.Digest()
			plan := BuildWithOperationInputs(strings.Repeat("a", 40), strings.Repeat("b", 40), bad,
				OperationInputs{WorkspaceSplits: map[string]workspacesplit.Request{modified.Digest(): modified}})
			if plan.Decision != DecisionRejected {
				t.Fatal("metric/request contradiction accepted", plan)
			}
			continue
		case "unreferenced":
			modified.Cases = "other.json"
			local.WorkspaceSplits[modified.Digest()] = modified
		}
		plan := BuildWithOperationInputs(strings.Repeat("a", 40), strings.Repeat("b", 40), report, local)
		if len(plan.Selected) != 0 || len(plan.OperationInputFailures) != 1 || plan.Reason != "OPERATION_INPUT_UNBOUND" {
			t.Fatal("unbound input acquired executable authority", mode, plan)
		}
	}
}
