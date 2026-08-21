package main

import "fmt"

func cycleIndicators(check validation, envelope Envelope, replay bool) []Indicator {
	return []Indicator{
		cycleBooleanIndicator("foundation.artifact-schemas", "FOUNDATION", check.Schemas),
		cycleBooleanIndicator("foundation.ci-run-binding", "FOUNDATION", check.Context),
		cycleBooleanIndicator("foundation.exact-head-binding", "FOUNDATION", check.Heads),
		cycleBooleanIndicator("foundation.gooo-contract-binding", "FOUNDATION", check.Contract),
		cycleBooleanIndicator("coherence.artifact-state", "COHERENCE", check.States),
		cycleBooleanIndicator("coherence.plan-execution-receipt-provenance", "COHERENCE", check.Links),
		cycleBooleanIndicator("coherence.indicator-ledger-binding", "COHERENCE", check.Ledger),
		cycleDigestIndicator("coherence.content-addressed-cycle", envelope.ArtifactSetDigest, check.Digests),
		cycleBooleanIndicator("foundation.source-metrics-schema", "FOUNDATION", check.MetricSchema),
		cycleBooleanIndicator("foundation.project-root-exemption", "FOUNDATION", check.MetricRootException),
		cycleBooleanIndicator("coherence.source-metrics-roots", "COHERENCE", check.MetricRoots),
		cycleBooleanIndicator("coherence.source-metrics-observations", "COHERENCE", check.MetricObservations),
		cycleDigestIndicator("coherence.source-metrics-witnesses", envelope.SourceMetrics.RootWitnessDigest, check.MetricWitnesses),
		cycleDigestIndicator("coherence.source-metrics-semantics", envelope.SourceMetrics.SemanticDigest, check.MetricSemantics),
		cycleBooleanIndicator("regression.canonical-replay", "REGRESSION", replay),
	}
}

func cycleBooleanIndicator(id, route string, pass bool) Indicator {
	return Indicator{ID: id, Route: route, Verdict: verdict(pass), Relation: "=", Value: fmt.Sprint(pass), Limit: "true"}
}

func cycleDigestIndicator(id, value string, pass bool) Indicator {
	return Indicator{ID: id, Route: "COHERENCE", Verdict: verdict(pass), Relation: "sha256", Value: value, Limit: "bound"}
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
