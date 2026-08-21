package generation

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestActionOutcomeValidationRejectsForgery(t *testing.T) {
	action := Action{
		MetricID:      "gooo.metric.source.file-lines.v1",
		Applicability: "APPLICABLE",
		Blocking:      true,
	}
	action.IndicatorOutcome = expectedActionOutcome(
		action.MetricID, action.Applicability, action.Blocking,
	)
	if err := validateActionOutcomes([]Action{action}); err != nil {
		t.Fatal(err)
	}
	if action.IndicatorOutcome.EnforcementEffect != sourcepolicy.EnforcementEffectBlock {
		t.Fatalf("outcome = %+v", action.IndicatorOutcome)
	}
	action.IndicatorOutcome.Decision = sourcepolicy.IndicatorDecisionPass
	if err := validateActionOutcomes([]Action{action}); err == nil {
		t.Fatal("forged action outcome accepted")
	}
}

func TestExecutionOutcomeValidationRejectsForgery(t *testing.T) {
	step := ExecutionStep{
		MetricID:      "gooo.metric.refactor.single-return.v1",
		Applicability: "APPLICABLE",
		Blocking:      false,
	}
	step.IndicatorOutcome = expectedActionOutcome(
		step.MetricID, step.Applicability, step.Blocking,
	)
	if err := validateExecutionOutcomes([]ExecutionStep{step}); err != nil {
		t.Fatal(err)
	}
	step.IndicatorOutcome.FailureCode = "forged"
	if err := validateExecutionOutcomes([]ExecutionStep{step}); err == nil {
		t.Fatal("forged execution outcome accepted")
	}
}
