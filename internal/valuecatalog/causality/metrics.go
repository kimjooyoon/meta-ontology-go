package causality

func deriveMetrics(graph GraphContract, resolutions []Resolution) Metrics {
	metrics := Metrics{
		ContractClaimTotal:   graph.NodeTotal,
		ContractEdgeTotal:    graph.EdgeTotal,
		ClassifiedClaimTotal: len(resolutions),
	}
	for _, resolution := range resolutions {
		switch resolution.Kind {
		case "EVIDENCE_ACCEPTED":
			metrics.DischargedClaimTotal++
		case "DIRECT_MISSING":
			metrics.UnknownClaimTotal++
			metrics.DirectMissingClaimTotal++
		case "DEPENDENCY_BLOCKED":
			metrics.UnknownClaimTotal++
			metrics.DependencyBlockedClaimTotal++
			metrics.ObservedBlockingEdgeTotal += len(resolution.BlockedByEdgeIDs)
		}
		depth := len(resolution.CausePath) - 1
		if depth > metrics.MaximumCausePathDepth {
			metrics.MaximumCausePathDepth = depth
		}
	}
	if graph.NodeTotal > 0 {
		metrics.ClassificationBasisPoints = metrics.ClassifiedClaimTotal * 10000 / graph.NodeTotal
		metrics.DischargeBasisPoints = metrics.DischargedClaimTotal * 10000 / graph.NodeTotal
	}
	return metrics
}

func buildIndicators(metrics Metrics) []Indicator {
	return []Indicator{
		{IndicatorID: "claim-contract-node-total", Class: "DRIVER", Trilemma: "FOUNDATION", Value: metrics.ContractClaimTotal, Target: ClaimTotal, Comparator: "EQ", Satisfied: metrics.ContractClaimTotal == ClaimTotal},
		{IndicatorID: "claim-contract-edge-total", Class: "DRIVER", Trilemma: "FOUNDATION", Value: metrics.ContractEdgeTotal, Target: EdgeTotal, Comparator: "EQ", Satisfied: metrics.ContractEdgeTotal == EdgeTotal},
		{IndicatorID: "claim-causally-classified-total", Class: "OUTCOME", Trilemma: "COHERENCE", Value: metrics.ClassifiedClaimTotal, Target: ClaimTotal, Comparator: "EQ", Satisfied: metrics.ClassifiedClaimTotal == ClaimTotal},
		{IndicatorID: "claim-direct-missing-total", Class: "GUARDRAIL", Trilemma: "REGRESSION", Value: metrics.DirectMissingClaimTotal, Target: 0, Comparator: "EQ", Satisfied: metrics.DirectMissingClaimTotal == 0},
		{IndicatorID: "claim-dependency-blocked-total", Class: "GUARDRAIL", Trilemma: "REGRESSION", Value: metrics.DependencyBlockedClaimTotal, Target: 0, Comparator: "EQ", Satisfied: metrics.DependencyBlockedClaimTotal == 0},
		{IndicatorID: "semantic-promotion-authority-total", Class: "GUARDRAIL", Trilemma: "FOUNDATION", Value: 0, Target: 0, Comparator: "EQ", Satisfied: true},
	}
}

func decisionFor(mode string, metrics Metrics) Decision {
	if mode == ModeSuccess && metrics.DischargedClaimTotal == ClaimTotal && metrics.UnknownClaimTotal == 0 {
		return Decision{Value: "PASS", Resolution: "CAUSAL_CLASSIFICATION_EXACT", Reason: "ALL_CLAIMS_EVIDENCE_ACCEPTED", SemanticPromotionAuthorized: false}
	}
	return Decision{Value: "FAIL_CLOSED", Resolution: "DEPENDENCY_LOCAL", Reason: "DIRECT_EVIDENCE_MISSING", SemanticPromotionAuthorized: false}
}

func receiptDigest(receipt Receipt) (string, error) {
	receipt.ReceiptDigest = ""
	return digestJSON(receipt)
}
