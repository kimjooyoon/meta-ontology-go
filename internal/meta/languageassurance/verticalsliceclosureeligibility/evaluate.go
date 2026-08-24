package verticalsliceclosureeligibility

func Evaluate(input Input) Report {
	report := baseReport(input)
	if !validSHA(input.SubjectSHA) || input.SubjectSHA == ShadowEvidenceHead {
		return lower(report, ResolutionUnknown, ReasonSubjectUnknown)
	}
	if len(input.Assurance.Payload) == 0 || len(input.Shadow.Payload) == 0 {
		return lower(report, ResolutionUnknown, ReasonUnavailable)
	}
	assurance, shadow, err := decodeEvidence(input)
	if err != nil {
		return lower(report, ResolutionInvariant, ReasonSemanticMismatch)
	}
	report.Summary.ObservedRepositoryWrites = shadow.RepositoryWrites
	if shadow.Decision != "SHADOW_PASS" && shadow.Decision != "FAIL_CLOSED" {
		return lower(report, ResolutionUnknown, ReasonDecisionUnknown)
	}
	if !evidenceLinked(assurance, shadow) {
		return lower(report, ResolutionInvariant, ReasonLinkMismatch)
	}
	if shadow.RepositoryWrites != 0 || shadow.PromotionApplied != 0 {
		return lower(report, ResolutionInvariant, ReasonWriteObserved)
	}
	if report.Summary.CapsulesExact != report.Summary.CapsulesTotal {
		return lower(report, ResolutionInvariant, ReasonDigestMismatch)
	}
	if !validAssurance(assurance) || !validShadow(shadow) {
		return lower(report, ResolutionInvariant, ReasonSemanticMismatch)
	}
	report.Decision, report.Resolution = DecisionEligible, ResolutionExact
	report.EnforcementEffect, report.Reason = EffectNone, ReasonEligible
	report.Summary.EligibleOperating, report.Summary.EligibleCoverageBPS = 11, 9166
	report.Summary.BoundariesTotal, report.Summary.BoundariesSatisfied = 6, 6
	report.Summary.LinksTotal, report.Summary.LinksSatisfied = 12, 12
	report.Summary.EligiblePaths, report.Summary.BlockedPaths = 1, 0
	return finish(report)
}

func baseReport(input Input) Report {
	artifacts := bindings(input)
	return Report{
		Schema: ReportSchema, SubjectSHA: input.SubjectSHA,
		AssuranceSubjectSHA: AssuranceEvidenceSubject, ShadowEvidenceHead: ShadowEvidenceHead,
		Decision: DecisionFailClosed, Resolution: ResolutionUnknown,
		EnforcementEffect: EffectBlock, Reason: ReasonUnavailable,
		AssuranceDenominatorID: AssuranceDenominatorID,
		AssuranceDenominatorDigest: AssuranceDenominatorDigest,
		ShadowDenominatorDigest: ShadowDenominatorDigest,
		Artifacts: artifacts, Transition: eligibilityTransition(), MetaOperations: MetaOperations(),
		Summary: Summary{DenominatorTotal: 12, BeforeOperating: 10, EligibleOperating: 10,
			OfficialOperating: 10, BeforeCoverageBPS: 8333, EligibleCoverageBPS: 8333,
			OfficialCoverageBPS: 8333, CapsulesTotal: 2, CapsulesExact: countExact(artifacts),
			CapsuleCoverageBPS: countExact(artifacts) * 5000, BlockedPaths: 1},
	}
}

func lower(report Report, resolution, reason string) Report {
	report.Resolution, report.Reason = resolution, reason
	if resolution == ResolutionUnknown {
		report.Summary.UnknownPaths = 1
	}
	return finish(report)
}

func finish(report Report) Report {
	report.Indicators = buildIndicators(report)
	return sealReport(report)
}

func eligibilityTransition() Transition {
	return Transition{MetricID: MetricID, MetaOperation: MetaOperation,
		FromStatus: "NOT_IMPLEMENTED", FromResolution: "NONE",
		EligibleStatus: "OPERATING", EligibleResolution: ResolutionExact,
		OfficialStatus: "NOT_IMPLEMENTED", OfficialResolution: "NONE"}
}
