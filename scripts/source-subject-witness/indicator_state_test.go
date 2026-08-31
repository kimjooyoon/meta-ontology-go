package main

import (
	"errors"
	"testing"
)

func TestSourceWitnessAcceptsUnsatisfiedLineDriversAsOpenObservations(t *testing.T) {
	for _, row := range []sourceIndicator{
		lineCapFixture(functionLinesMetric, "fixture.go:3:Large", 90, 75, false),
		lineCapFixture("gooo.metric.source.go-file-lines.v1", "fixture.go", 80, 75, false),
		lineCapFixture("gooo.metric.source.gooo-file-lines.v1", "fixture.gooo", 80, 75, false),
	} {
		if err := validateIndicatorShape(row); err != nil {
			t.Fatalf("line-cap shape rejected: %v", err)
		}
		if err := validateIndicatorState(row); err != nil {
			t.Fatalf("unsatisfied line-cap observation rejected: %v", err)
		}
	}
	observations := unsatisfiedLineDrivers([]sourceIndicator{
		lineCapFixture(functionLinesMetric, "fixture.go:3:Large", 90, 75, false),
		lineCapFixture("gooo.metric.source.go-file-lines.v1", "fixture.go", 80, 75, false),
	})
	if len(observations) != 2 || observationState(observations) != "OBSERVED" || claimState(observations) != "OPEN" {
		t.Fatalf("source observation state = %q/%q with %d rows", observationState(observations), claimState(observations), len(observations))
	}
}

func TestSourceWitnessRejectsCrossMetricLineDriverProducers(t *testing.T) {
	for _, row := range []sourceIndicator{
		func() sourceIndicator {
			row := lineCapFixture(functionLinesMetric, "fixture.go:3:Large", 90, 75, false)
			row.Producer = "linecaps.AnalyzeLineMetrics"
			return row
		}(),
		func() sourceIndicator {
			row := lineCapFixture("gooo.metric.source.go-file-lines.v1", "fixture.go", 80, 75, false)
			row.Producer = "linecaps.Analyze"
			return row
		}(),
		func() sourceIndicator {
			row := lineCapFixture("gooo.metric.source.gooo-file-lines.v1", "fixture.gooo", 80, 75, false)
			row.Producer = "linecaps.Analyze"
			return row
		}(),
	} {
		err := validateLineCapIndicator(row)
		if err == nil {
			t.Fatalf("cross-metric producer %q was accepted for %s", row.Producer, row.MetricID)
		}
		assertSourceValidation(t, err, "REFUTED", "INVARIANT_ONLY", "", "report-counterexample")
		var validationErr *sourceValidationError
		if !errors.As(err, &validationErr) || validationErr.Reason != "SOURCE_LINE_CAP_DRIVER_CONTRADICTION" {
			t.Fatalf("cross-metric producer reason = %v", err)
		}
	}
}

func TestSourceWitnessAcceptsSatisfiedLineDriver(t *testing.T) {
	row := lineCapFixture(functionLinesMetric, "fixture.go:3:Small", 75, 75, true)
	if err := validateIndicatorState(row); err != nil {
		t.Fatalf("satisfied line-cap observation rejected: %v", err)
	}
}

func TestSourceWitnessRejectsMalformedAndBlockingLineDrivers(t *testing.T) {
	if err := validateIndicatorShape(sourceIndicator{MetricID: functionLinesMetric}); err == nil {
		t.Fatal("incomplete indicator was accepted")
	} else {
		assertSourceValidation(t, err, "FAIL_CLOSED", "LOWER_RESOLUTION", "MALFORMED_EVIDENCE", "restore-source-metrics")
	}

	row := lineCapFixture(functionLinesMetric, "fixture.go:3:Large", 90, 75, false)
	row.Blocking = true
	err := validateIndicatorState(row)
	if err == nil {
		t.Fatal("blocking line-cap contradiction was accepted")
	}
	assertSourceValidation(t, err, "REFUTED", "INVARIANT_ONLY", "", "report-counterexample")
}

func lineCapFixture(metric, subject string, value, limit int, satisfied bool) sourceIndicator {
	contract, _ := lineCapContract(metric)
	row := sourceIndicator{
		Applicability:       "APPLICABLE",
		ApplicabilityReason: "CATALOG_APPLICABLE",
		ApplicabilityRuleID: defaultApplicabilityRule,
		Blocking:            false,
		Consumer:            contract.consumer,
		Decision:            "FAIL_CLOSED",
		EnforcementEffect:   "NO_EFFECT",
		EvaluationState:     "EVALUATED",
		FailureCode:         metric + "#predicate-false",
		FailureReason:       "PREDICATE_FALSE",
		Family:              contract.family,
		Limit:               limit,
		MetaOperation:       contract.operation,
		MetricID:            metric,
		Producer:            contract.producer,
		ProofChoice:         "foundation",
		Relation:            "less_or_equal",
		Role:                "DRIVER",
		Satisfied:           satisfied,
		Subject:             subject,
		SubjectKind:         contract.kind,
		Value:               value,
	}
	if satisfied {
		row.Decision = "PASS"
		row.FailureCode = ""
		row.FailureReason = "NONE"
	}
	return row
}

func assertSourceValidation(t *testing.T, err error, decision, resolution, unknownClass, next string) {
	t.Helper()
	var validationErr *sourceValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("error type = %T, want source validation error", err)
	}
	if validationErr.Decision != decision || validationErr.Resolution != resolution || validationErr.UnknownClass != unknownClass || validationErr.NextOperation != next || len(validationErr.BlockedBy) != 0 {
		t.Fatalf("source validation = %#v", validationErr)
	}
}
