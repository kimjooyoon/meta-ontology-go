package toolchainformatfix

func finish(report Report, cycle cycleEvidence, registryDrift int, unknownReason string) Report {
	report.Summary = summarize(report.Source, report.Cases, cycle, registryDrift)
	report.RepositoryWrites = report.Summary.RepositoryWrites
	report.Decision, report.Resolution = DecisionClosed, ResolutionExact
	report.ReasonCode = "FORMAT_FIX_CASE_NOT_SATISFIED"
	if report.Summary.Unresolved > 0 {
		report.Resolution, report.ReasonCode = ResolutionLower, unknownReason
	}
	report.Indicators = indicators(report.Summary, report.Resolution)
	report.Proofs = proofs(report, cycle)
	if report.Summary.Satisfied == FixedTotal && allIndicators(report.Indicators) &&
		allProofs(report.Proofs) {
		report.Decision, report.ReasonCode = DecisionPass, "ALL_FORMAT_FIX_CASES_SATISFIED"
	}
	return seal(report)
}

func allIndicators(indicators []Indicator) bool {
	for _, indicator := range indicators {
		if !indicator.Satisfied {
			return false
		}
	}
	return true
}

func allProofs(proofs []Proof) bool {
	for _, proof := range proofs {
		if !proof.Passed {
			return false
		}
	}
	return true
}
