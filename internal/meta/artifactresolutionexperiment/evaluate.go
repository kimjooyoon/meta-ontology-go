package artifactresolutionexperiment

func Evaluate(input Input) Report {
	if reason := validate(input); reason != "" {
		return closed(input, "EXACT", reason)
	}
	if topDecisionUnknown(input) {
		return closed(input, "LOWER_RESOLUTION", "ARTIFACT_DECISION_UNKNOWN")
	}
	facts := digestValue(input)
	indicators := buildIndicators(input)
	summary := summarize(input, indicators)
	report := Report{Schema: ReportSchema, Decision: "PASS", Resolution: "EXACT",
		Reason: "ARTIFACT_RESOLUTION_OBSERVED", Interpretation: "RESOLUTION_EXPANSION_OBSERVED",
		SubjectSHA: input.SubjectSHA, ContractID: input.Contract.ID,
		Summary: summary, Indicators: indicators, Views: buildViews(indicators),
		Proofs: buildProofs(indicators, facts), NotClaimed: input.Contract.NotClaimed,
		FactsDigest: facts}
	if summary.Coordinates.Satisfied != summary.Coordinates.Total {
		report.Decision = "FAIL_CLOSED"
		report.Reason = "ARTIFACT_RESOLUTION_CONTRACT_NOT_SATISFIED"
		report.Interpretation = "NO_LANGUAGE_QUALITY_CLAIM"
	}
	return seal(report)
}
