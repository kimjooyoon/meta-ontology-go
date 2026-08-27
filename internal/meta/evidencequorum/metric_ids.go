package evidencequorum

func MetricIDs() []string {
	return []string{
		"gooo.metric.meta.evidence-quorum.fixed-case-coverage.v1",
		"gooo.metric.meta.evidence-quorum.independent-quorum.v1",
		"gooo.metric.meta.evidence-quorum.duplicate-not-independent.guardrail.v1",
		"gooo.metric.meta.evidence-quorum.conflict-fail-closed.guardrail.v1",
		"gooo.metric.meta.evidence-quorum.insufficient-lower-resolution.guardrail.v1",
		"gooo.metric.meta.evidence-quorum.confidence-aggregation.guardrail.v1",
		"gooo.metric.meta.evidence-quorum.claim-transitions.v1",
		"gooo.metric.meta.evidence-quorum.observer-writes.guardrail.v1",
	}
}
