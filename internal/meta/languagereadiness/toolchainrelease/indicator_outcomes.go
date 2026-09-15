package toolchainrelease

func outcomeIndicators(summary Summary) []Indicator {
	proofBPS := 10000
	if summary.ProofFailures != 0 {
		proofBPS = 0
	}
	return []Indicator{
		indicator(outcomeMetricIDs[0], "OUTCOME", "COHERENCE", summary.ReadinessBPS, 10000, "greater_or_equal"),
		indicator(outcomeMetricIDs[1], "OUTCOME", "COHERENCE", summary.ReadinessBPS, 10000, "greater_or_equal"),
		indicator(outcomeMetricIDs[2], "OUTCOME", "REGRESSION", proofBPS, 10000, "greater_or_equal"),
	}
}
