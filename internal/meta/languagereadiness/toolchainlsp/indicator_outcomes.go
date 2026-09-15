package toolchainlsp

func outcomeIndicators(summary Summary, resolution string) []Indicator {
	proofReadiness := 0
	if summary.ProofFailures == 0 {
		proofReadiness = 10000
	}
	return []Indicator{
		indicator("readiness-bps.v1", "OUTCOME", "COHERENCE", summary.ReadinessBPS, 10000, "greater_or_equal", resolution),
		indicator("case-readiness-bps.v1", "OUTCOME", "COHERENCE", summary.ReadinessBPS, 10000, "greater_or_equal", resolution),
		indicator("proof-readiness-bps.v1", "OUTCOME", "REGRESSION", proofReadiness, 10000, "greater_or_equal", resolution),
	}
}
