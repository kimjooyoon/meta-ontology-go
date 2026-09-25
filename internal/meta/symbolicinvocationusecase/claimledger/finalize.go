package claimledger

func finalize(report *Report, expected ExpectedMetrics) {
	report.Metrics.FixedClaimTotal = len(report.Claims)
	report.Metrics.OpenClaimTotal = len(report.OpenClaimIDs)
	if report.Metrics.InScopeClaimTotal > 0 {
		report.Metrics.DischargeBasisPoints = report.Metrics.DischargedTotal * 10_000 / report.Metrics.InScopeClaimTotal
	}
	switch {
	case report.Metrics.InScopeClaimTotal == 0:
		report.ClaimSet = Verdict{"FAIL_CLOSED", "NONE", "NO_SCOPED_CLAIMS"}
	case report.Metrics.UnknownTotal > 0:
		report.ClaimSet = Verdict{"FAIL_CLOSED", "STAGE_LOCAL", "OPEN_UNKNOWN_CLAIMS_REMAIN"}
	case report.Metrics.RefutedTotal > 0:
		report.ClaimSet = Verdict{"FAIL_CLOSED", "CLAIM_LOCAL", "REFUTED_CLAIMS_PRESENT"}
	default:
		report.ClaimSet = Verdict{"PASS", "EXACT", "ALL_SCOPED_CLAIMS_DISCHARGED"}
	}
	if report.ClaimSet.Decision == "PASS" && (report.Metrics.UnknownTotal > 0 || report.Metrics.RefutedTotal > 0) {
		report.Metrics.FalsePromotionCount = 1
	}
	if metricsMatch(report, expected) {
		report.Conformance = Conformance{"PASS", "FIXED_METRIC_CONTRACT_MATCHED"}
	} else {
		report.Conformance = Conformance{"FAIL_CLOSED", "FIXED_METRIC_CONTRACT_MISMATCH"}
	}
}

func metricsMatch(report *Report, expected ExpectedMetrics) bool {
	m := report.Metrics
	return m.FixedClaimTotal == expected.FixedClaimTotal &&
		m.InScopeClaimTotal == expected.InScopeClaimTotal &&
		m.DischargedTotal == expected.DischargedTotal &&
		m.UnknownTotal == expected.UnknownTotal &&
		m.RefutedTotal == expected.RefutedTotal &&
		m.ExcludedTotal == expected.ExcludedTotal &&
		m.OpenClaimTotal == expected.OpenClaimTotal &&
		m.DischargeBasisPoints == expected.DischargeBasisPoints &&
		m.FalsePromotionCount == expected.FalsePromotionCount &&
		m.ProofRoutes == expected.ProofRoutes &&
		report.ClaimSet.Decision == expected.ClaimSetDecision &&
		report.ClaimSet.Resolution == expected.Resolution
}
