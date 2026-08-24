package rollbackintegrityshadow

func Evaluate(raw []byte) Report {
	assurance, resolution, reason := inspectAssurance(raw)
	if reason != "" {
		return failureReport(raw, resolution, reason)
	}
	cases, summary := runSuite()
	summary.DenominatorTotal, summary.BeforeOperating = 12, 9
	summary.BeforeCoverageBPS = 7500
	summary.ProjectedOperating, summary.ProjectedCoverageBPS = 9, 7500
	exact := summary.CasesPassed == caseTotal && summary.MetaReportsValid == caseTotal &&
		summary.CoordinatesTotal == caseTotal*10 && summary.UnknownDecisionCases == 1
	if exact {
		summary.ProjectedOperating, summary.ProjectedCoverageBPS = 10, 8333
	}
	report := Report{Schema: Schema, MetricID: MetricID, MetaOperation: MetaOperation,
		Decision: DecisionFailClosed, Reason: ReasonSuite, Resolution: ResolutionInvariant,
		EnforcementEffect: EnforcementNoEffect, AssuranceSubjectSHA: assurance.SubjectSHA,
		EvidenceDigest: digestBytes(raw), Summary: summary, Cases: cases,
		RepositoryWrites: 0, PromotionApplied: 0}
	if exact {
		report.Decision, report.Reason, report.Resolution =
			DecisionShadowPass, ReasonShadowPass, ResolutionExact
	}
	report.Indicators = buildIndicators(summary, true)
	return seal(report)
}

func failureReport(raw []byte, resolution, reason string) Report {
	summary := Summary{DenominatorTotal: 12, BeforeOperating: 9, ProjectedOperating: 9,
		BeforeCoverageBPS: 7500, ProjectedCoverageBPS: 7500}
	report := Report{Schema: Schema, MetricID: MetricID, MetaOperation: MetaOperation,
		Decision: DecisionFailClosed, Reason: reason, Resolution: resolution,
		EnforcementEffect: EnforcementNoEffect, EvidenceDigest: digestBytes(raw),
		Summary: summary, RepositoryWrites: 0, PromotionApplied: 0}
	report.Indicators = buildIndicators(summary, false)
	return seal(report)
}
