package selectiveci

func sealCoverage(result ObligationCoverageResult, decision CoverageDecision, reason CoverageReason) ObligationCoverageResult {
	result.Decision = decision
	result.Reason = reason
	result.FullSuiteRequired = decision == CoverageDecisionUnknown
	if decision == CoverageDecisionUnknown {
		result.RequiredObligationIDs = []string{}
	}
	result = normalizeCoverageResult(result)
	result.OutputDigest = result.StableDigest()
	return result
}

func normalizeCoverageResult(result ObligationCoverageResult) ObligationCoverageResult {
	result.UncoveredRootIDs = sortedUnique(result.UncoveredRootIDs)
	result.RequiredObligationIDs = sortedUnique(result.RequiredObligationIDs)
	if result.UncoveredRootIDs == nil {
		result.UncoveredRootIDs = []string{}
	}
	if result.RequiredObligationIDs == nil {
		result.RequiredObligationIDs = []string{}
	}
	return result
}
