package languagedeterministicquery

func newSource(input Input, registryDigest string) Source {
	return Source{
		ExpectedHeadSHA: input.ExpectedHeadSHA,
		ConceptID: ConceptID,
		Producer: "languagedeterministicquery.Evaluate",
		Consumer: "self-improvement-cycle",
		MetaOperation: ExpectedMetaOperation,
		ConceptArtifactDigest: input.ConceptArtifact.ArtifactDigest,
		CatalogDigest: input.ConceptArtifact.CatalogDigest,
		RegistryDigest: registryDigest,
	}
}

func successOrFailureReport(source Source, summary Summary, results []CaseResult) Report {
	report := Report{
		Schema: ReportSchema, Decision: DecisionFailClosed, Resolution: ResolutionLower,
		ReasonCode: "QUERY_CASES_NOT_SATISFIED", Source: source, Summary: summary, Cases: results,
		RepositoryWrites: 0, MutationAuthorized: false,
	}
	report.Stages = stages(summary)
	report.Indicators = indicators(summary, report.Resolution)
	report.Proofs = proofs(summary, source.RegistryDigest)
	if summary.Satisfied == FixedTotal && summary.Executed == FixedTotal &&
		allStagesPassed(report.Stages) && allProofsPassed(report.Proofs) {
		report.Decision = DecisionPass
		report.Resolution = ResolutionExact
		report.ReasonCode = "ALL_FIXED_QUERY_CASES_SATISFIED"
		report.Indicators = indicators(summary, report.Resolution)
	}
	return finalizeReport(report)
}

func failureReport(source Source, reason string) Report {
	summary := Summary{Total: FixedTotal, Unresolved: FixedTotal, RegistryDrift: 1}
	report := Report{
		Schema: ReportSchema, Decision: DecisionFailClosed, Resolution: ResolutionLower,
		ReasonCode: reason, Source: source, Summary: summary,
		RepositoryWrites: 0, MutationAuthorized: false,
	}
	report.Stages = stages(summary)
	report.Indicators = indicators(summary, report.Resolution)
	report.Proofs = proofs(summary, source.RegistryDigest)
	return finalizeReport(report)
}

func finalizeReport(report Report) Report {
	report.ReportDigest = ""
	report.ReportDigest = digestJSON(report)
	return report
}
