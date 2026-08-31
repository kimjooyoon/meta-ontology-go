package verticalsliceclosureshadow

func Evaluate(input Input) Report {
	return evaluate(input, activeDenominator())
}

func evaluate(input Input, denominatorRaw []byte) Report {
	report := baseReport(input, denominatorRaw)
	value, err := decodeDenominator(denominatorRaw)
	if err != nil {
		return finishFailure(report, ResolutionInvariant, ReasonDenominator)
	}
	subject, resolution, reason := inspectAssurance(input.Assurance)
	report.AssuranceSubjectSHA = subject
	if reason != "" {
		return finishFailure(report, resolution, reason)
	}
	report.Boundaries = make([]BoundaryResult, len(value.Boundaries))
	artifacts := make(map[string]artifactEnvelope, len(value.Boundaries))
	results := make(map[string]*BoundaryResult, len(value.Boundaries))
	for index, spec := range value.Boundaries {
		result, artifact := inspectBoundary(spec, input.Artifacts[spec.ID], input.HeadSHA)
		report.Boundaries[index], artifacts[spec.ID] = result, artifact
		results[spec.ID] = &report.Boundaries[index]
	}
	for _, spec := range value.Boundaries {
		applyBoundaryLinks(results[spec.ID], artifacts[spec.ID], artifacts,
			input.Artifacts, results, input.HeadSHA)
	}
	report.Summary = summarize(report.Boundaries)
	report.Decision, report.Reason, report.Resolution =
		DecisionFailClosed, ReasonBoundaryBlocked, ResolutionInvariant
	if report.Summary.UnknownBoundaries > 0 {
		report.Reason, report.Resolution = ReasonEvidenceUnknown, ResolutionLower
	} else if report.Summary.BoundariesSatisfied == boundaryTotal &&
		report.Summary.LinksSatisfied == linkTotal {
		report.Decision, report.Reason, report.Resolution =
			DecisionShadowPass, ReasonShadowPass, ResolutionExact
		report.Summary.ProjectedOperating = projectedOperating
		report.Summary.ProjectedCoverageBPS = projectedCoverageBPS
	}
	report.Indicators = buildIndicators(report.Summary)
	return seal(report)
}

func baseReport(input Input, denominatorRaw []byte) Report {
	summary := baselineSummary()
	return Report{Schema: Schema, MetricID: MetricID, MetaOperation: MetaOperation,
		Decision: DecisionFailClosed, Reason: ReasonBoundaryBlocked,
		Resolution: ResolutionInvariant, EnforcementEffect: EnforcementNoEffect,
		HeadSHA: input.HeadSHA, AssuranceDigest: digestBytes(input.Assurance),
		DenominatorDigest: digestBytes(denominatorRaw), Summary: summary,
		RepositoryWrites: 0, PromotionApplied: 0}
}

func finishFailure(report Report, resolution, reason string) Report {
	report.Decision, report.Resolution, report.Reason = DecisionFailClosed, resolution, reason
	report.Indicators = buildIndicators(report.Summary)
	return seal(report)
}
