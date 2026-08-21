package main

import "fmt"

func cycleIndicators(check validation, envelope Envelope, replay bool) []Indicator {
	return []Indicator{
		{ID: "foundation.artifact-schemas", Route: "FOUNDATION",
			Verdict: verdict(check.Schemas), Relation: "=", Value: fmt.Sprint(check.Schemas), Limit: "true"},
		{ID: "foundation.ci-run-binding", Route: "FOUNDATION",
			Verdict: verdict(check.Context), Relation: "=", Value: fmt.Sprint(check.Context), Limit: "true"},
		{ID: "foundation.exact-head-binding", Route: "FOUNDATION",
			Verdict: verdict(check.Heads), Relation: "=", Value: fmt.Sprint(check.Heads), Limit: "true"},
		{ID: "foundation.gooo-contract-binding", Route: "FOUNDATION",
			Verdict: verdict(check.Contract), Relation: "=", Value: fmt.Sprint(check.Contract), Limit: "true"},
		{ID: "coherence.artifact-state", Route: "COHERENCE",
			Verdict: verdict(check.States), Relation: "=", Value: fmt.Sprint(check.States), Limit: "true"},
		{ID: "coherence.plan-execution-receipt-provenance", Route: "COHERENCE",
			Verdict: verdict(check.Links), Relation: "=", Value: fmt.Sprint(check.Links), Limit: "true"},
		{ID: "coherence.indicator-ledger-binding", Route: "COHERENCE",
			Verdict: verdict(check.Ledger), Relation: "=", Value: fmt.Sprint(check.Ledger), Limit: "true"},
			{ID: "coherence.content-addressed-cycle", Route: "COHERENCE",
				Verdict: verdict(check.Digests), Relation: "sha256",
				Value: envelope.ArtifactSetDigest, Limit: "bound"},
			{ID: "coherence.source-metrics-semantics", Route: "COHERENCE",
				Verdict: verdict(check.Metrics), Relation: "sha256",
				Value: envelope.SourceMetrics.SemanticDigest, Limit: "bound"},
			{ID: "regression.canonical-replay", Route: "REGRESSION",
			Verdict: verdict(replay), Relation: "=", Value: fmt.Sprint(replay), Limit: "true"},
	}
}

func finishEnvelope(envelope *Envelope) {
	envelope.Status, envelope.Reason = "BOUND", "SELF_IMPROVEMENT_CYCLE_BOUND"
	for _, indicator := range envelope.Indicators {
		if indicator.Verdict != "PASS" {
			envelope.Status, envelope.Reason = "OPEN", "SELF_IMPROVEMENT_CYCLE_OPEN"
			return
		}
	}
}

func verdict(pass bool) string {
	if pass {
		return "PASS"
	}
	return "FAIL"
}

func metricExpectations(metrics MetricsBinding) map[string]int {
	return map[string]int{
		"gooo.metric.layout.direct-files.v1":      metrics.StorageRoot.DirectFiles,
		"gooo.metric.layout.direct-folders.v1":    metrics.StorageRoot.DirectFolders,
		"gooo.metric.layout.recursive-files.v1":   metrics.StorageRoot.RecursiveFiles,
		"gooo.metric.layout.recursive-folders.v1": metrics.StorageRoot.RecursiveFolders,
		"gooo.metric.source.go-files.v1":          metrics.LogicalRoot.GoFiles,
		"gooo.metric.source.go-lines.v1":          metrics.LogicalRoot.GoLines,
		"gooo.metric.source.gooo-files.v1":        metrics.LogicalRoot.GoooFiles,
		"gooo.metric.source.gooo-lines.v1":        metrics.LogicalRoot.GoooLines,
	}
}

func rootException(indicator metricsIndicator, root MetricsSnapshot) bool {
	expected := root.DirectFolders + root.DirectFiles
	if indicator.MetricID == "gooo.metric.layout.entry-kinds.v1" {
		expected = 0
		if root.DirectFolders > 0 {
			expected++
		}
		if root.DirectFiles > 0 {
			expected++
		}
	} else if indicator.MetricID != "gooo.metric.layout.direct-entries.v1" {
		return false
	}
	return indicator.Value == expected && indicator.Applicability == "NOT_APPLICABLE" &&
		indicator.ApplicabilityReason == "ROOT_TOPOLOGY_EXEMPT" &&
		!indicator.Blocking && indicator.Decision == "NOT_APPLICABLE"
}
