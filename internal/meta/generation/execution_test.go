package generation

import (
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestBuildExecutionManifestIsDeterministicAndSandboxed(t *testing.T) {
	plan := actionableReceiptPlan()
	first := BuildExecutionManifest(plan)
	replay := BuildExecutionManifest(plan)
	if !reflect.DeepEqual(first, replay) {
		t.Fatal("execution manifest did not replay deterministically")
	}
	if first.Decision != ExecutionDecisionProposed ||
		first.Reason != ExecutionReasonIndependentActions ||
		len(first.Steps) != 2 || first.ReplayDigest == "" {
		t.Fatalf("unexpected execution manifest: %+v", first)
	}
	if first.PromotionAuthorized ||
		first.PromotionAuthorizedByExecution() {
		t.Fatal("execution manifest acquired promotion authority")
	}
	for _, step := range first.Steps {
		if step.WorkspaceMode != WorkspaceModeDisposable ||
			step.WriteBoundary != WriteBoundarySandboxOnly ||
			!step.ReceiptRequired || step.Activity == "" || step.Output == "" ||
			len(step.RequiredIndicatorIDs) == 0 {
			t.Fatalf("unsafe execution step: %+v", step)
		}
	}
}

func TestBuildExecutionManifestFailsClosedAndFindsFixedPoint(t *testing.T) {
	report := sourcepolicy.Report{
		Schema: sourcepolicy.IndicatorSchema, Policy: sourcepolicy.Default(),
		Indicators: []sourcepolicy.Indicator{
			metric("floor", sourcepolicy.OperationSplitGo, true, true),
		},
	}
	plan := Build(strings.Repeat("5", 40), strings.Repeat("6", 40), report)
	fixed := BuildExecutionManifest(plan)
	if fixed.Decision != ExecutionDecisionFixedPoint ||
		fixed.Reason != ExecutionReasonExactFixedPoint ||
		len(fixed.Steps) != 0 {
		t.Fatalf("unexpected fixed point: %+v", fixed)
	}
	plan.PlanDigest = strings.Repeat("0", 64)
	unknown := BuildExecutionManifest(plan)
	if unknown.Decision != ExecutionDecisionUnknown ||
		unknown.Reason != ExecutionReasonInvalidPlan {
		t.Fatalf("tampered plan did not fail closed: %+v", unknown)
	}
}
