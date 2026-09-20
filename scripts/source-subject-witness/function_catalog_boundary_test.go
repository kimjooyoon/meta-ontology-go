package main

import "testing"

func TestCompileFunctionWitnessesPreservesCatalogBoundary(t *testing.T) {
	supported := functionFixture()
	outOfCatalog := supported
	outOfCatalog.MetricID = "gooo.metric.refactor.duplicate-body.v1"
	outOfCatalog.Relation = "less_or_equal"
	outOfCatalog.Value = 1
	outOfCatalog.Limit = 0
	outOfCatalog.Blocking = true
	outOfCatalog.EnforcementEffect = "BLOCK"
	outOfCatalog.FailureCode = outOfCatalog.MetricID + "#predicate-false"
	outOfCatalog.FailureReason = "PREDICATE_FALSE"
	outOfCatalog.Decision = "FAIL_CLOSED"

	witnesses, err := compileFunctionWitnesses([]sourceIndicator{supported, outOfCatalog})
	if err != nil {
		t.Fatalf("out-of-catalog function indicator escaped the catalog boundary: %v", err)
	}
	if len(witnesses) != 1 || witnesses[0].Path != supported.Subject {
		t.Fatalf("catalog witness set changed unexpectedly: %#v", witnesses)
	}
}
