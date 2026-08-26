package candidateleakage

func Evaluate(input Input) Report {
	denominator := DenominatorContract()
	report := Report{
		Schema: ReportSchema, SubjectSHA: input.SubjectSHA,
		Decision: DecisionFailClosed, Resolution: ResolutionInvariant,
		EnforcementEffect: EffectBlock, Reason: ReasonDecisionUnknown,
		DenominatorID: denominator.ID, DenominatorDigest: denominator.Digest,
		Input: input, MetaOperations: MetaOperations(),
		Summary: Summary{BoundaryPaths: 1, BlockedPaths: 1},
	}
	if reason := validateBoundary(input); reason != "" {
		report.Reason = reason
		report.Summary.UnknownPaths = 1
		return finish(report)
	}
	if input.Promotion.CandidateDigest != input.Candidate.Digest {
		report.Reason = ReasonDigestBindingMismatch
		report.Summary.UnknownPaths = 1
		return finish(report)
	}
	if !sameOperation(input) {
		report.Reason = ReasonOperationBindingMismatch
		report.Summary.UnknownPaths = 1
		return finish(report)
	}
	authorized := input.Promotion.Decision == PromotionAuthorized
	positive := positiveOfficial(input.Official)
	report.Resolution = ResolutionExact
	report.Summary.BoundaryBindingBPS = 10_000
	report.Summary.PromotionAuthorityBPS = 10_000
	if positive && !authorized {
		report.Reason = ReasonLeakageDetected
		report.Summary.LeakagePaths = 1
		return finish(report)
	}
	report.Decision = DecisionPass
	report.EnforcementEffect = EffectNone
	report.Summary.BlockedPaths = 0
	if authorized && positive {
		report.Reason = ReasonExactPromotionBound
		report.Summary.AuthorizedPaths = 1
	} else {
		report.Reason = ReasonCandidateIsolated
	}
	return finish(report)
}

func finish(report Report) Report {
	report.Indicators = buildIndicators(report.Summary, report.Resolution)
	return sealReport(report)
}

func sameOperation(input Input) bool {
	operation := input.Candidate.MetaOperation
	return operation == BoundaryOperation && input.Promotion.MetaOperation == operation &&
		input.Official.MetaOperation == operation
}

func positiveOfficial(official Official) bool {
	if official.Status == OfficialOperating {
		return true
	}
	switch official.Decision {
	case OfficialAllow, OfficialPass, OfficialFixedPoint, OfficialAuthorized:
		return true
	default:
		return false
	}
}
