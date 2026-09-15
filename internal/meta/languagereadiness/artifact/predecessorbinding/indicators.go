package predecessorbinding

func indicators(report Report) []Indicator {
	summary := report.Summary
	return []Indicator{
		newIndicator("gooo.metric.language.predecessor-dynamic-binding-bps.v1", "outcome",
			"COHERENCE", "classify-predecessor-bindings", summary.DynamicBPS, 10_000, "BASIS_POINT"),
		newIndicator("gooo.metric.language.predecessor-dynamic-coordinates.v1", "driver",
			"COHERENCE", "count-dynamic-predecessor-coordinates", summary.DynamicInput, Total, "COORDINATE"),
		newIndicator("gooo.metric.language.predecessor-static-coordinates.guardrail.v1", "guardrail",
			"REGRESSION", "count-static-predecessor-coordinates", summary.StaticLiteral, 0, "COORDINATE"),
		newIndicator("gooo.metric.language.predecessor-unknown-coordinates.guardrail.v1", "guardrail",
			"FOUNDATION", "lower-resolution-on-unknown-coordinate", summary.Unknown, 0, "COORDINATE"),
		newIndicator("gooo.metric.language.predecessor-observer-writes.guardrail.v1", "guardrail",
			"FOUNDATION", "preserve-read-only-observation", report.RepositoryWrites, 0, "REPOSITORY_WRITE"),
	}
}

func newIndicator(metricID, class, proof, operation string, value, target int, unit string) Indicator {
	satisfied := value >= target
	if class == "guardrail" {
		satisfied = value <= target
	}
	return Indicator{MetricID: metricID, Class: class, ProofChoice: proof,
		Producer: "predecessorbinding.Evaluate", Consumer: "self-improvement-cycle",
		MetaOperation: operation, Value: value, Target: target, Unit: unit,
		Satisfied: satisfied}
}

func proofs(report Report) []Proof {
	return []Proof{
		{ID: "fixed-eight-coordinate-registry", Choice: "FOUNDATION", Passed: len(coordinates) == Total},
		{ID: "ast-binding-classification", Choice: "COHERENCE", Passed: report.Summary.Unknown == 0},
		{ID: "unknown-fails-closed", Choice: "REGRESSION", Passed: report.Summary.Unknown == 0},
		{ID: "read-only-observer", Choice: "FOUNDATION", Passed: report.RepositoryWrites == 0},
	}
}
