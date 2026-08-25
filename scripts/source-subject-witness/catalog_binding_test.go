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

func TestRootSummaryIndicatorsJoinTheRootBinding(t *testing.T) {
	directory, rows := rootFixture()
	binding, err := storageDirectoryBinding(directory, indexIndicators(rows), 0)
	if err != nil {
		t.Fatal(err)
	}
	if binding.IndicatorCount != 12 || binding.ApplicableIndicators != 9 || binding.NotApplicableIndicators != 3 {
		t.Fatalf("root binding applicability counts = %d/%d/%d", binding.IndicatorCount, binding.ApplicableIndicators, binding.NotApplicableIndicators)
	}
	rows[len(rows)-1].MetaOperation = "unknown"
	if _, err := storageDirectoryBinding(directory, indexIndicators(rows), 0); err == nil {
		t.Fatal("unknown root meta operation was accepted")
	}
}

func functionFixture() sourceIndicator {
	return sourceIndicator{Subject: "main.go:1:main", SubjectKind: "FUNCTION", MetricID: functionMetric, Value: 1, Relation: "observe", Applicability: "APPLICABLE", ApplicabilityRuleID: defaultApplicabilityRule, ApplicabilityReason: "CATALOG_APPLICABLE", Satisfied: true, ProofChoice: "coherence", Producer: "linecaps.Analyze", Consumer: "refactor-report", MetaOperation: "inspect-wrapper", Detail: "single return CallExpr", Decision: "PASS", EvaluationState: "EVALUATED", FailureReason: "NONE", EnforcementEffect: "NO_EFFECT"}
}

func rootFixture() (directoryMetric, []sourceIndicator) {
	directory := directoryMetric{Path: ".", SubjectKind: "PROJECT_ROOT", DirectFolders: 1, DirectFiles: 2, RecursiveFolders: 1, RecursiveFiles: 2, GoFiles: 1, GoooFiles: 1, GoLines: 10, GoooLines: 5}
	metrics := []expectedMetric{{"gooo.metric.layout.direct-entries.v1", 3}, {"gooo.metric.layout.direct-files.v1", 2}, {"gooo.metric.layout.direct-folders.v1", 1}, {"gooo.metric.layout.entry-kinds.v1", 2}, {"gooo.metric.layout.recursive-files.v1", 2}, {"gooo.metric.layout.recursive-folders.v1", 1}, {rootREADMEMetric, 0}}
	metrics = append(metrics, rootSummaryMetrics(directory)...)
	rows := make([]sourceIndicator, 0, len(metrics))
	for _, metric := range metrics {
		rows = append(rows, sourceIndicator{Subject: ".", SubjectKind: "PROJECT_ROOT", MetricID: metric.id, Value: metric.value, Relation: "observe", Applicability: "APPLICABLE", ApplicabilityRuleID: defaultApplicabilityRule, ApplicabilityReason: "CATALOG_APPLICABLE", Satisfied: true, ProofChoice: "coherence", Producer: "linecaps.AnalyzeLineMetrics", Consumer: "metric-report", MetaOperation: "observe", Decision: "PASS", EvaluationState: "EVALUATED", FailureReason: "NONE", EnforcementEffect: "NO_EFFECT"})
	}
	for _, index := range []int{0, 3, 6} {
		rows[index].Applicability, rows[index].Decision = "NOT_APPLICABLE", "NOT_APPLICABLE"
	}
	rows[6].Detail = "ontology=" + rootREADMEOntology
	last := len(rows) - 1
	rows[last].Relation, rows[last].Producer, rows[last].Consumer, rows[last].MetaOperation, rows[last].Detail, rows[last].Blocking = "less_or_equal", "metabinding.Build", "self-improvement-cycle", "bind-indicator-meta-program", "ontology=examples/meta-binding-coverage/main.gooo", true
	return directory, rows
}
