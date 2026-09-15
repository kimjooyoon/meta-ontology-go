package rollbackintegrityeligibility

func Evaluate(input Input) Report {
	denominator := denominatorContract()
	report := Report{
		Schema: ReportSchema, SubjectSHA: input.SubjectSHA, EvidenceSubjectSHA: EvidenceSubjectSHA,
		Decision: DecisionFailClosed, Resolution: ResolutionUnknown,
		EnforcementEffect: EffectBlock, Reason: ReasonUnavailable,
		DenominatorID: DenominatorID, DenominatorDigest: denominator.DenominatorDigest,
		Artifacts: bindings(input), Transition: eligibilityTransition(), MetaOperations: MetaOperations(),
		RepositoryWrites: 0, PromotionApplied: 0,
		Summary: Summary{DenominatorTotal: 12, BeforeOperating: 9, AfterOperating: 9,
			BeforeCoverageBPS: 7500, AfterCoverageBPS: 7500, CapsulesTotal: 3, BlockedPaths: 1},
	}
	if !validSHA(input.SubjectSHA) || input.SubjectSHA == EvidenceSubjectSHA {
		report.Reason = ReasonSubjectUnknown
		report.Summary.UnknownPaths = 1
		return finish(report)
	}
	if len(input.Assurance.Payload) == 0 || len(input.ShadowReportA.Payload) == 0 ||
		len(input.ShadowReportB.Payload) == 0 {
		report.Summary.UnknownPaths = 1
		return finish(report)
	}
	if !exactDigests(input) {
		report.Resolution, report.Reason = ResolutionInvariant, ReasonDigestMismatch
		report.Summary.UnknownPaths = 1
		return finish(report)
	}
	assurance, assuranceErr := decode[assuranceCapsule](input.Assurance.Payload)
	shadowA, reportAErr := decode[shadowReportCapsule](input.ShadowReportA.Payload)
	shadowB, reportBErr := decode[shadowReportCapsule](input.ShadowReportB.Payload)
	if assuranceErr != nil || reportAErr != nil || reportBErr != nil ||
		!validSemantics(assurance, shadowA, shadowB) {
		report.Resolution, report.Reason = ResolutionInvariant, ReasonSemanticMismatch
		report.Summary.UnknownPaths = 1
		return finish(report)
	}
	report.Decision, report.Resolution = DecisionEligible, ResolutionExact
	report.EnforcementEffect, report.Reason = EffectNone, ReasonEligible
	report.Summary = Summary{DenominatorTotal: 12, BeforeOperating: 9, AfterOperating: 10,
		BeforeCoverageBPS: 7500, AfterCoverageBPS: 8333,
		CapsulesTotal: 3, CapsulesExact: 3, CapsuleCoverageBPS: 10_000,
		ShadowCasesTotal: 7, ShadowCasesPassed: 7,
		ShadowReplaysTotal: 2, ShadowReplaysExact: 2, ShadowReplayCoverageBPS: 10_000,
		MetaOperationsRequired: 1, MetaOperationsObserved: 1, MetaOperationCoverageBPS: 10_000,
		EligiblePaths: 1}
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
