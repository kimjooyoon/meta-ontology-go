package generation

import (
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestPlanAndExecutionBindMetricApplicability(t *testing.T) {
	report, err := sourcepolicy.Evaluate(sourcepolicy.Default(), []sourcepolicy.Observation{
		{Subject: ".", Dimension: sourcepolicy.DimensionDirectEntries, Value: 99},
		{Subject: ".", Dimension: sourcepolicy.DimensionDirectoryKinds, Value: 2},
		{Subject: ".", Dimension: sourcepolicy.DimensionRootREADME, Value: 0},
		{Subject: ".github/workflows", Dimension: sourcepolicy.DimensionDirectEntries, Value: 17,
			Detail: sourcepolicy.WorkflowDiscoveryObservationDetail},
	})
	if err != nil {
		t.Fatal(err)
	}
	report.Indicators = append(report.Indicators,
		metric("expression", sourcepolicy.OperationCollapseAssign, false, false),
		metric("topology", sourcepolicy.OperationSplitGo, false, false),
	)
	plan := Build(strings.Repeat("7", 40), strings.Repeat("8", 40), report)
	if plan.Decision != DecisionPlan || len(plan.Selected) != 2 ||
		len(plan.NotApplicableIndicatorIDs) != 4 {
		t.Fatalf("applicability was not represented in plan: %+v", plan)
	}
	for _, action := range plan.Selected {
		if !validActionApplicability(action) {
			t.Fatalf("action lost metric applicability: %+v", action)
		}
	}
	manifest := BuildExecutionManifest(plan)
	if manifest.Decision != ExecutionDecisionProposed ||
		len(manifest.NotApplicableIndicatorIDs) != 4 || len(manifest.Steps) != 2 {
		t.Fatalf("execution lost applicability evidence: %+v", manifest)
	}
	for _, step := range manifest.Steps {
		if step.Applicability != sourcepolicy.ApplicabilityApplicable ||
			step.ApplicabilityRule != sourcepolicy.ApplicabilityRuleDefault ||
			step.MetricProofChoice == "" || step.MetricConsumer == "" {
			t.Fatalf("execution step lost metric binding: %+v", step)
		}
	}
}

func TestWorkflowDiscoveryRootApplicabilityIsAccepted(t *testing.T) {
	indicator := sourcepolicy.Indicator{
		MetricID: sourcepolicy.DimensionDirectEntries, Family: sourcepolicy.FamilyTopology,
		Subject: ".github/workflows", SubjectKind: sourcepolicy.SubjectKindDirectory,
		Value: 36, Limit: 0, Relation: sourcepolicy.RelationObserve,
		Applicability:       sourcepolicy.ApplicabilityNotApplicable,
		ApplicabilityRule:   sourcepolicy.ApplicabilityRuleWorkflowDiscoveryRoot,
		ApplicabilityReason: sourcepolicy.ApplicabilityReasonWorkflowRootExempt,
		Satisfied:           true, Proof: sourcepolicy.ProofFoundation,
		Producer: "repository-projector.topology", Consumer: "github-actions",
		Operation: sourcepolicy.OperationExemptWorkflowRoot,
	}
	if !validIndicatorApplicability(indicator) {
		t.Fatalf("workflow discovery root exemption was rejected: %+v", indicator)
	}
}
