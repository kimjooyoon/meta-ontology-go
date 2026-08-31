package candidateleakageeligibility

func Evaluate(input Input) Report {
	denominator := denominatorContract()
	report := Report{
		Schema: ReportSchema, SubjectSHA: input.SubjectSHA, EvidenceSubjectSHA: EvidenceSubjectSHA,
		Decision: DecisionFailClosed, Resolution: ResolutionUnknown,
		EnforcementEffect: EffectBlock, Reason: ReasonUnavailable,
		DenominatorID: DenominatorID, DenominatorDigest: denominator.DenominatorDigest,
		Artifacts: bindings(input), Transition: eligibilityTransition(),
		MetaOperations: MetaOperations(), RepositoryWrites: 0, PromotionApplied: 0,
		Summary: Summary{DenominatorTotal: 12, BeforeOperating: 7, AfterOperating: 7,
			BeforeCoverageBPS: 5833, AfterCoverageBPS: 5833, CapsulesTotal: 2, BlockedPaths: 1},
	}
	if !validSHA(input.SubjectSHA) || input.SubjectSHA == EvidenceSubjectSHA {
		report.Reason = ReasonSubjectUnknown
		report.Summary.UnknownPaths = 1
		return finish(report)
	}
	if len(input.Assurance.Payload) == 0 || len(input.Shadow.Payload) == 0 {
		report.Summary.UnknownPaths = 1
		return finish(report)
	}
	if !exactDigests(input) {
		report.Resolution, report.Reason = ResolutionInvariant, ReasonDigestMismatch
		report.Summary.UnknownPaths = 1
		return finish(report)
	}
	assurance, assuranceErr := decodeAssurance(input.Assurance.Payload)
	shadow, shadowErr := decodeShadow(input.Shadow.Payload)
	if assuranceErr != nil || shadowErr != nil || !validSemantics(assurance, shadow) {
		report.Resolution, report.Reason = ResolutionInvariant, ReasonSemanticMismatch
		report.Summary.UnknownPaths = 1
		return finish(report)
	}
	report.Decision, report.Resolution = DecisionEligible, ResolutionExact
	report.EnforcementEffect, report.Reason = EffectNone, ReasonEligible
	report.Summary = Summary{
		DenominatorTotal: 12, BeforeOperating: 7, AfterOperating: 8,
		BeforeCoverageBPS: 5833, AfterCoverageBPS: 6666,
		CapsulesTotal: 2, CapsulesExact: 2, CapsuleCoverageBPS: 10_000,
		ShadowCasesTotal: 6, ShadowCasesPassed: 6, EligiblePaths: 1,
	}
	return finish(report)
}

func finish(report Report) Report {
	report.Indicators = buildIndicators(report.Summary, report.Resolution)
	return sealReport(report)
}

func eligibilityTransition() Transition {
	return Transition{MetricID: MetricID, MetaOperation: MetaOperation,
		FromStatus: "NOT_IMPLEMENTED", FromResolution: "NONE",
		EligibleStatus: "OPERATING", EligibleResolution: "EXACT"}
}
