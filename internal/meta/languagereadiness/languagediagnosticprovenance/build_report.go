package languagediagnosticprovenance

func successOrFailureReport(source Source, summary Summary, results []CaseResult) Report {
	report := Report{
		Schema: ReportSchema, Decision: DecisionFailClosed,
		Resolution: ResolutionLower,
		ReasonCode: "DIAGNOSTIC_PROVENANCE_CASES_NOT_SATISFIED",
		Source:     source, Summary: summary, Cases: results,
		RepositoryWrites: 0, MutationAuthorized: false,

		Stages: stages(summary)}
	report.Indicators = indicators(summary, report.Resolution)
	report.Proofs = proofs(summary, source.RegistryDigest)
	if exactSummary(summary) && allStagesPassed(report.Stages) &&
		allProofsPassed(report.Proofs) {
		report.Decision, report.Resolution = DecisionPass, ResolutionExact
		report.ReasonCode = "ALL_FIXED_DIAGNOSTIC_PROVENANCE_CASES_SATISFIED"
		report.Indicators = indicators(summary, report.Resolution)
	}
	return finalizeReport(report)
}

func failureReport(source Source, reason string) Report {
	summary := Summary{
		Total: FixedTotal, Unresolved: FixedTotal,
		RegistryDrift: 1,
	}
	report := Report{
		Schema: ReportSchema, Decision: DecisionFailClosed,
		Resolution: ResolutionLower, ReasonCode: reason,
		Source: source, Summary: summary,
		RepositoryWrites: 0, MutationAuthorized: false,

		Stages: stages(summary)}
	report.Indicators = indicators(summary, report.Resolution)
	report.Proofs = proofs(summary, source.RegistryDigest)
	return finalizeReport(report)
}

func finalizeReport(report Report) Report {
	report.ReportDigest = ""
	report.ReportDigest = digestJSON(report)
	return report
}

func exactSummary(summary Summary) bool {
	return summary.Satisfied == 18 && summary.Total == 18 && summary.Executed == 18 &&
		summary.Traced == 10 && summary.GuardrailRejections == 8 &&
		summary.PhysicalPositions == 10 && summary.LogicalPositions == 10 &&
		summary.SemanticBindings == 4 && summary.LSPProjections == 10 &&
		summary.CanonicalReplays == 10 && summary.OrderedDiagnostics == 6 &&
		summary.LineDirectiveRemaps == 1 && summary.TypeClassifications == 3 &&
		summary.ProvenanceSteps == 50 && summary.ConceptBindings == 1 &&
		summary.CodeBindings == 8 && summary.MetricBindings == 18 &&
		summary.UseCaseBindings == 3 && summary.ToolchainMatches == 1 &&
		summary.NotSatisfied == 0 && summary.Unresolved == 0 &&
		summary.RegistryDrift == 0 && summary.ConceptDrift == 0
}
