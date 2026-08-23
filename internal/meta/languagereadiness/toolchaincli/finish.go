package toolchaincli

func finish(report Report, registryDrift int, unknownReason string) Report {
	report.Summary = summarize(report.Source, report.Cases, registryDrift)
	report.RepositoryWrites = report.Summary.RepositoryWrites
	report.Decision, report.Resolution = DecisionClosed, ResolutionExact
	report.ReasonCode = "TOOLCHAIN_CLI_CASE_NOT_SATISFIED"
	if report.Summary.Unresolved > 0 {
		report.Resolution = ResolutionLower
		report.ReasonCode = unknownReason
		if report.ReasonCode == "" {
			report.ReasonCode = "TOOLCHAIN_CLI_OBSERVATION_UNKNOWN"
		}
	}
	report.Indicators = indicators(report.Summary, report.Resolution, report.MutationAuthorized)
	report.Proofs = proofs(report)
	report.Stages = stages(report)
	if report.Summary.Satisfied == FixedTotal && allIndicators(report.Indicators) && allProofs(report.Proofs) {
		report.Decision, report.ReasonCode = DecisionPass, "ALL_TOOLCHAIN_CLI_CASES_SATISFIED"
	}
	return seal(report)
}
