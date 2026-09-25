package generation

import (
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestGenerationRejectsUnprovenOrExecutableExemptions(t *testing.T) {
	invalid := metric("invalid", sourcepolicy.OperationCollapseAssign, false, false)
	invalid.Applicability = sourcepolicy.ApplicabilityNotApplicable
	report := sourcepolicy.Report{
		Schema: sourcepolicy.IndicatorSchema, Policy: sourcepolicy.Default(),
		Indicators: []sourcepolicy.Indicator{invalid},
	}
	plan := Build(strings.Repeat("9", 40), strings.Repeat("a", 40), report)
	if plan.Decision != DecisionUnknown || plan.Reason != ReasonApplicabilityUnproven {
		t.Fatalf("unproven applicability did not fail closed: %+v", plan)
	}

	valid := actionableReceiptPlan()
	valid.Selected[0].Applicability = sourcepolicy.ApplicabilityNotApplicable
	valid = finish(valid)
	manifest := BuildExecutionManifest(valid)
	if manifest.Decision != ExecutionDecisionUnknown ||
		manifest.Reason != ExecutionReasonInvalidPlan {
		t.Fatalf("executable exemption did not fail closed: %+v", manifest)
	}
}
