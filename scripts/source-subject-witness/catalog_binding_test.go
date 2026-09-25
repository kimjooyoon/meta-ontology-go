package main

import "testing"

func TestFunctionIndicatorsCompileIntoMetaBoundWitnesses(t *testing.T) {
	row := functionFixture()
	witnesses, err := compileFunctionWitnesses([]sourceIndicator{row})
	if err != nil {
		t.Fatal(err)
	}
	if len(witnesses) != 1 || witnesses[0].Space != "LOGICAL_FUNCTION" || witnesses[0].Meta.IndicatorCount != 1 || witnesses[0].Meta.ApplicableIndicators != 1 {
		t.Fatalf("function witnesses = %#v", witnesses)
	}
}

func TestFunctionIndicatorsRejectUnknownCatalogEntries(t *testing.T) {
	for _, mutate := range []func(*sourceIndicator){
		func(row *sourceIndicator) { row.MetricID = "unknown" },
		func(row *sourceIndicator) { row.MetaOperation = "unknown" },
	} {
		row := functionFixture()
		mutate(&row)
		if _, err := compileFunctionWitnesses([]sourceIndicator{row}); err == nil {
			t.Fatal("unknown function catalog entry was accepted")
		}
	}
}

func TestFunctionIndicatorsShareWitnessAcrossMetrics(t *testing.T) {
	singleReturn := functionFixture()
	lineCap := lineCapFixture(functionLinesMetric, singleReturn.Subject, 42, 75, true)
	witnesses, err := compileFunctionWitnesses([]sourceIndicator{singleReturn, lineCap})
	if err != nil {
		t.Fatal(err)
	}
	if len(witnesses) != 1 || witnesses[0].Meta.IndicatorCount != 2 || len(witnesses[0].Metrics) != 2 {
		t.Fatalf("function witnesses = %#v", witnesses)
	}
	if witnesses[0].Metrics[0].ID != functionMetric || witnesses[0].Metrics[1].ID != functionLinesMetric {
		t.Fatalf("function metric order = %#v", witnesses[0].Metrics)
	}
}

func TestFunctionIndicatorsRejectDuplicateMetricForSubject(t *testing.T) {
	row := functionFixture()
	if _, err := compileFunctionWitnesses([]sourceIndicator{row, row}); err == nil {
		t.Fatal("duplicate function metric was accepted")
	}
}

func functionFixture() sourceIndicator {
	return sourceIndicator{Subject: "main.go:1:main", SubjectKind: "FUNCTION", MetricID: functionMetric, Value: 1, Relation: "observe", Applicability: "APPLICABLE", ApplicabilityRuleID: defaultApplicabilityRule, ApplicabilityReason: "CATALOG_APPLICABLE", Satisfied: true, ProofChoice: "coherence", Producer: "linecaps.Analyze", Consumer: "refactor-report", MetaOperation: "inspect-wrapper", Detail: "single return CallExpr", Decision: "PASS", EvaluationState: "EVALUATED", FailureReason: "NONE", EnforcementEffect: "NO_EFFECT"}
}
