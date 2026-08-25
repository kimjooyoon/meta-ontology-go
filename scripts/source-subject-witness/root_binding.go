package main

import "fmt"

const (
	rootUnboundMetric = "gooo.metric.meta.unbound-indicators.v1"
	rootSummaryCount  = 5
)

func rootSummaryMetrics(directory directoryMetric) []expectedMetric {
	return []expectedMetric{
		{"gooo.metric.source.go-files.v1", directory.GoFiles},
		{"gooo.metric.source.go-lines.v1", directory.GoLines},
		{"gooo.metric.source.gooo-files.v1", directory.GoooFiles},
		{"gooo.metric.source.gooo-lines.v1", directory.GoooLines},
		{rootUnboundMetric, 0},
	}
}

func validateRootSummaryIndicator(row sourceIndicator) error {
	if !isRootSummaryMetric(row.MetricID) {
		return fmt.Errorf("root summary indicator %q is outside the exact catalog", row.MetricID)
	}
	wantOperation, wantProducer, wantConsumer, wantEffect, wantBlocking := "observe", "linecaps.AnalyzeLineMetrics", "metric-report", "NO_EFFECT", false
	if row.MetricID == rootUnboundMetric {
		wantOperation, wantProducer, wantConsumer, wantEffect, wantBlocking = "bind-indicator-meta-program", "metabinding.Build", "self-improvement-cycle", "ALLOW", true
		if row.Value != 0 || row.Limit != 0 || row.Relation != "less_or_equal" || row.Detail != "ontology=examples/meta-binding-coverage/main.gooo" {
			return fmt.Errorf("root unbound indicator lost its exact zero-value ontology binding")
		}
	} else if row.Value < 0 || row.Limit != 0 || row.Relation != "observe" {
		return fmt.Errorf("root source observation %q lost its non-negative observation contract", row.MetricID)
	}
	if row.Subject != "." || row.SubjectKind != "PROJECT_ROOT" || row.MetaOperation != wantOperation || row.Producer != wantProducer || row.Consumer != wantConsumer || row.ProofChoice != "coherence" || row.EnforcementEffect != wantEffect || row.Blocking != wantBlocking || !exactApplicable(row) {
		return fmt.Errorf("root summary indicator %q is outside the exact catalog", row.MetricID)
	}
	return nil
}

func isRootSummaryMetric(metric string) bool {
	for _, item := range rootSummaryMetrics(directoryMetric{}) {
		if item.id == metric {
			return true
		}
	}
	return false
}
