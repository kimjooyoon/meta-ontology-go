package main

import "testing"

func TestRootSummaryCompilesInLogicalCoordinateSpace(t *testing.T) {
	storage, root, rows := rootFixture()
	index := indexIndicators(rows)
	storageBinding, err := storageDirectoryBinding(storage, index, 0)
	if err != nil {
		t.Fatal(err)
	}
	if storageBinding.IndicatorCount != 7 || storageBinding.ApplicableIndicators != 4 || storageBinding.NotApplicableIndicators != 3 {
		t.Fatalf("storage root applicability = %d/%d/%d", storageBinding.IndicatorCount, storageBinding.ApplicableIndicators, storageBinding.NotApplicableIndicators)
	}
	summary, err := compileRootSummaryWitness(root, index)
	if err != nil {
		t.Fatal(err)
	}
	if summary.Space != rootSummarySpace || summary.Meta.IndicatorCount != 5 || summary.Meta.ApplicableIndicators != 5 || summary.Metrics[4].Value != root.GoFiles || root.GoFiles == storage.GoFiles {
		t.Fatalf("logical source summary = %#v", summary)
	}
}

func TestRootSummaryRejectsCoordinateAndCatalogMismatch(t *testing.T) {
	_, root, rows := rootFixture()
	root.GoFiles++
	if _, err := compileRootSummaryWitness(root, indexIndicators(rows)); err == nil {
		t.Fatal("logical source coordinate mismatch was accepted")
	}
	root.GoFiles--
	rows[len(rows)-1].MetaOperation = "unknown"
	if _, err := compileRootSummaryWitness(root, indexIndicators(rows)); err == nil {
		t.Fatal("unknown root meta operation was accepted")
	}
}

func rootFixture() (directoryMetric, directoryMetric, []sourceIndicator) {
	storage := directoryMetric{Path: ".", SubjectKind: "PROJECT_ROOT", DirectFolders: 1, DirectFiles: 2, RecursiveFolders: 1, RecursiveFiles: 2, GoFiles: 1, GoooFiles: 1, GoLines: 10, GoooLines: 5}
	root := directoryMetric{Path: ".", SubjectKind: "PROJECT_ROOT", DirectFolders: 7, DirectFiles: 7, RecursiveFolders: 8, RecursiveFiles: 9, GoFiles: 101, GoooFiles: 11, GoLines: 1001, GoooLines: 111}
	metrics := []expectedMetric{{"gooo.metric.layout.direct-entries.v1", 3}, {"gooo.metric.layout.direct-files.v1", 2}, {"gooo.metric.layout.direct-folders.v1", 1}, {"gooo.metric.layout.entry-kinds.v1", 2}, {"gooo.metric.layout.recursive-files.v1", 2}, {"gooo.metric.layout.recursive-folders.v1", 1}, {rootREADMEMetric, 0}}
	metrics = append(metrics, rootSummaryMetrics(root)...)
	rows := make([]sourceIndicator, 0, len(metrics))
	for _, metric := range metrics {
		rows = append(rows, sourceIndicator{Subject: ".", SubjectKind: "PROJECT_ROOT", MetricID: metric.id, Value: metric.value, Relation: "observe", Applicability: "APPLICABLE", ApplicabilityRuleID: defaultApplicabilityRule, ApplicabilityReason: "CATALOG_APPLICABLE", Satisfied: true, ProofChoice: "coherence", Producer: "linecaps.AnalyzeLineMetrics", Consumer: "metric-report", MetaOperation: "observe", Decision: "PASS", EvaluationState: "EVALUATED", FailureReason: "NONE", EnforcementEffect: "NO_EFFECT"})
	}
	for _, index := range []int{0, 3, 6} {
		rows[index].Applicability, rows[index].Decision = "NOT_APPLICABLE", "NOT_APPLICABLE"
	}
	rows[6].Detail = "ontology=" + rootREADMEOntology
	last := len(rows) - 1
	rows[last].Relation, rows[last].Producer, rows[last].Consumer, rows[last].MetaOperation, rows[last].Detail, rows[last].Blocking, rows[last].EnforcementEffect = "less_or_equal", "metabinding.Build", "self-improvement-cycle", "bind-indicator-meta-program", "ontology=examples/meta-binding-coverage/main.gooo", true, "ALLOW"
	return storage, root, rows
}
