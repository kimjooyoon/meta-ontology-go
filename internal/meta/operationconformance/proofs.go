package operationconformance

func buildProofs(report Report) []Proof {
	foundation := routePassed(report.Indicators, "FOUNDATION") && report.Contract.Decision == DecisionPass
	coherence := routePassed(report.Indicators, "COHERENCE")
	regression := routePassed(report.Indicators, "REGRESSION")
	return []Proof{
		{Choice: "FOUNDATION", MetaOperation: "bind-contract-source-and-build-context",
			Passed: foundation, EvidenceDigest: report.EvidenceDigest},
		{Choice: "COHERENCE", MetaOperation: "cohere-write-set-and-candidates",
			Passed: coherence, EvidenceDigest: report.EvidenceDigest},
		{Choice: "REGRESSION", MetaOperation: "reject-header-effects-and-unknown",
			Passed: regression, EvidenceDigest: report.EvidenceDigest},
	}
}

func routePassed(indicators []IndicatorObservation, route string) bool {
	matched := false
	for _, indicator := range indicators {
		if indicator.Route != route {
			continue
		}
		matched = true
		if indicator.Decision != DecisionPass {
			return false
		}
	}
	return matched
}
