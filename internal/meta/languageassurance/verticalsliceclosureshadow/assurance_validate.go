package verticalsliceclosureshadow

func validAssurance(report assuranceReport) bool {
	summary := report.Summary
	return report.Schema == "gooo/language-assurance-report/v1" &&
		report.SubjectSHA == PredecessorSHA && report.AssuranceDecision == "PARTIAL" &&
		report.CandidateDecision == "ALLOW_LIMITED" && summary.DenominatorTotal == officialTotal &&
		summary.Operating == beforeOperating && summary.NotImplemented == 2 &&
		summary.CoverageBPS == beforeCoverageBPS && summary.UnknownTopDecisions == 0 &&
		summary.UnresolvedIndicators == 0 && summary.ViolatedGuardrails == 0 &&
		summary.RepositoryWrites == 0 && validTargetObligation(report.Obligations)
}

func validTargetObligation(obligations []assuranceObligation) bool {
	matches := 0
	for _, obligation := range obligations {
		if obligation.MetricID != MetricID {
			continue
		}
		matches++
		if obligation.Status != "NOT_IMPLEMENTED" || obligation.Resolution != "NONE" ||
			obligation.MetaOperation != "" {
			return false
		}
	}
	return matches == 1
}
