package toolchainconformance

func buildIndicators(summary Summary) []Indicator {
	indicators := outcomeIndicators(summary)
	indicators = append(indicators, driverIndicators(summary)...)
	return append(indicators, guardrailIndicators(summary)...)
}

func outcomeIndicators(summary Summary) []Indicator {
	return []Indicator{
		metric("gooo.metric.toolchain.conformance-readiness-bps.v1",
			"OUTCOME", "COHERENCE", summary.ReadinessBPS, 10000, "greater_or_equal"),
		metric("gooo.metric.toolchain.conformance-case-readiness-bps.v1",
			"OUTCOME", "COHERENCE", summary.CaseReadinessBPS, 10000, "greater_or_equal"),
		metric("gooo.metric.toolchain.conformance-proof-readiness-bps.v1",
			"OUTCOME", "REGRESSION", summary.ProofReadinessBPS, 10000, "greater_or_equal"),
	}
}
