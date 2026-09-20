package main

import "testing"

func TestValidateIndicatorStatePreservesExplicitUnsatisfiedNonBlockingIndicator(t *testing.T) {
	row := sourceIndicator{
		MetricID:            "gooo.metric.refactor.assign-return.v1",
		Applicability:       "APPLICABLE",
		ApplicabilityReason: "CATALOG_APPLICABLE",
		ApplicabilityRuleID: defaultApplicabilityRule,
		Blocking:            false,
		Decision:            "FAIL_CLOSED",
		EvaluationState:     "EVALUATED",
		EnforcementEffect:   "NO_EFFECT",
		FailureCode:         "gooo.metric.refactor.assign-return.v1#predicate-false",
		FailureReason:       "PREDICATE_FALSE",
		Relation:            "equal",
		Value:               2,
		Limit:               0,
	}
	if err := validateIndicatorState(row); err != nil {
		t.Fatalf("explicit non-blocking counterexample was rejected: %v", err)
	}
	observations := unsatisfiedIndicators([]sourceIndicator{row})
	if len(observations) != 1 || observations[0].MetricID != row.MetricID {
		t.Fatalf("explicit counterexample was not preserved: %#v", observations)
	}
}

func TestValidateIndicatorStatePreservesExplicitBlockingIndicator(t *testing.T) {
	row := sourceIndicator{
		MetricID:            "gooo.metric.refactor.duplicate-body.v1",
		Applicability:       "APPLICABLE",
		ApplicabilityReason: "CATALOG_APPLICABLE",
		ApplicabilityRuleID: defaultApplicabilityRule,
		Blocking:            true,
		Decision:            "FAIL_CLOSED",
		EvaluationState:     "EVALUATED",
		EnforcementEffect:   "BLOCK",
		FailureCode:         "gooo.metric.refactor.duplicate-body.v1#predicate-false",
		FailureReason:       "PREDICATE_FALSE",
		Relation:            "less_or_equal",
		Value:               1,
		Limit:               0,
	}
	if err := validateIndicatorState(row); err != nil {
		t.Fatalf("explicit blocking counterexample was rejected: %v", err)
	}
}
