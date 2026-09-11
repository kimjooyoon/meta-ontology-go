package generation

import (
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestActionOutcomeValidationRejectsForgery(t *testing.T) {
	indicator := sourcepolicy.Indicator{
		MetricID:      "gooo.metric.source.file-lines.v1",
		SubjectKind:   sourcepolicy.SubjectKindFile,
		Applicability: "APPLICABLE",
		Blocking:      true,
	}
	action := Action{
		IndicatorID: indicatorID(indicator), MetricID: indicator.MetricID,
		SubjectKind: indicator.SubjectKind, InputSubjectKind: indicator.SubjectKind,
		InputContractSourceDigest: strings.Repeat("a", 64), InputContractSemanticDigest: strings.Repeat("b", 64),
		Applicability: indicator.Applicability, Blocking: indicator.Blocking,
		SourceIndicator: indicator, IndicatorOutcome: indicator.Outcome(),
	}
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
	indicator := sourcepolicy.Indicator{
		MetricID:      "gooo.metric.refactor.single-return.v1",
		SubjectKind:   sourcepolicy.SubjectKindFunction,
		Applicability: "APPLICABLE",
		Blocking:      false,
	}
	step := ExecutionStep{
		ActionIndicatorID: indicatorID(indicator), MetricID: indicator.MetricID,
		SubjectKind: indicator.SubjectKind, InputSubjectKind: indicator.SubjectKind,
		InputContractSourceDigest: strings.Repeat("a", 64), InputContractSemanticDigest: strings.Repeat("b", 64),
		Applicability: indicator.Applicability, Blocking: indicator.Blocking,
		SourceIndicator: indicator, IndicatorOutcome: indicator.Outcome(),
	}
	if err := validateExecutionOutcomes([]ExecutionStep{step}); err != nil {
		t.Fatal(err)
	}
	step.IndicatorOutcome.FailureCode = "forged"
	if err := validateExecutionOutcomes([]ExecutionStep{step}); err == nil {
		t.Fatal("forged execution outcome accepted")
	}
}
