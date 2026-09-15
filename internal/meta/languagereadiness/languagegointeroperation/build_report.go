package languagegointeroperation

func successOrFailureReport(source Source, summary Summary, results []CaseResult) Report {
	report := Report{Schema: ReportSchema, Decision: DecisionFailClosed, Resolution: ResolutionLower,
		ReasonCode: "GO_INTEROPERATION_CASES_NOT_SATISFIED", Source: source, Summary: summary,
		Cases: results, RepositoryWrites: 0, MutationAuthorized: false, Stages: stages(summary)}
	report.Indicators = indicators(summary, report.Resolution)
	report.Proofs = proofs(summary, source.RegistryDigest)
	if exactSummary(summary) && allStagesPassed(report.Stages) && allProofsPassed(report.Proofs) {
		report.Decision, report.Resolution = DecisionPass, ResolutionExact
		report.ReasonCode = "ALL_FIXED_GO_INTEROPERATION_CASES_SATISFIED"
		report.Indicators = indicators(summary, report.Resolution)
	}
	return finalizeReport(report)
}

func failureReport(source Source, reason string) Report {
	summary := Summary{Total: FixedTotal, Unresolved: FixedTotal, RegistryDrift: 1}
	report := Report{Schema: ReportSchema, Decision: DecisionFailClosed, Resolution: ResolutionLower,
		ReasonCode: reason, Source: source, Summary: summary, RepositoryWrites: 0,
		MutationAuthorized: false, Stages: stages(summary)}
	report.Indicators = indicators(summary, report.Resolution)
	report.Proofs = proofs(summary, source.RegistryDigest)
	return finalizeReport(report)
}

func exactSummary(summary Summary) bool {
	return summary.Satisfied == 24 && summary.Executed == 24 && summary.GeneratorProjections == 8 &&
		summary.Go127Boundaries == 8 && summary.GuardrailRejections == 8 && summary.PositiveAccepted == 16 &&
		summary.CanonicalReplays == 16 && summary.TypeIdentityReplays == 16 && summary.SourceMaps == 8 &&
		summary.GenericMethods == 5 && summary.AliasNodes == 2 && summary.ASTReifications == 32 &&
		summary.NotSatisfied == 0 && summary.Unresolved == 0 && summary.RegistryDrift == 0
}

func finalizeReport(report Report) Report {
	report.ReportDigest = ""
	report.ReportDigest = digestJSON(report)
	return report
}
