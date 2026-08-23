package toolchainconformance

func driverIndicators(summary Summary) []Indicator {
	return []Indicator{
		metric("gooo.metric.toolchain.conformance-surfaces.v1", "DRIVER",
			"FOUNDATION", summary.SurfacesSatisfied, ExpectedSurfaceCount, "greater_or_equal"),
		metric("gooo.metric.toolchain.conformance-executed-cases.v1", "DRIVER",
			"COHERENCE", summary.ExecutedCases, ExpectedCaseCount, "greater_or_equal"),
		metric("gooo.metric.toolchain.conformance-source-indicators.v1", "DRIVER",
			"COHERENCE", summary.IndicatorsSatisfied, ExpectedIndicatorCount, "greater_or_equal"),
		metric("gooo.metric.toolchain.conformance-source-proofs.v1", "DRIVER",
			"REGRESSION", summary.ProofsPassed, ExpectedProofCount, "greater_or_equal"),
		metric("gooo.metric.toolchain.conformance-head-bindings.v1", "DRIVER",
			"COHERENCE", summary.HeadBindings, ExpectedSurfaceCount, "greater_or_equal"),
		metric("gooo.metric.toolchain.conformance-tamper-rejections.v1", "DRIVER",
			"REGRESSION", summary.TamperRejections, ExpectedTamperCount, "greater_or_equal"),
		metric("gooo.metric.toolchain.conformance-concept-bindings.v1", "DRIVER",
			"FOUNDATION", summary.ConceptBindings, 1, "greater_or_equal"),
		metric("gooo.metric.toolchain.conformance-code-bindings.v1", "DRIVER",
			"FOUNDATION", summary.CodeBindings, 5, "greater_or_equal"),
		metric("gooo.metric.toolchain.conformance-metric-bindings.v1", "DRIVER",
			"COHERENCE", summary.MetricBindings, ExpectedMetricCount, "greater_or_equal"),
		metric("gooo.metric.toolchain.conformance-use-case-bindings.v1", "DRIVER",
			"REGRESSION", summary.UseCaseBindings, 3, "greater_or_equal"),
	}
}
