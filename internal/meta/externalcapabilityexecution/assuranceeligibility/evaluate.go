package assuranceeligibility

func Evaluate(input Input) Report {
	report := baseReport(input)
	if !validSHA(input.SubjectSHA) || input.SubjectSHA == AssuranceSubject {
		return lower(report, ResolutionUnknown, ReasonSubjectUnknown)
	}
	if !available(input) {
		return lower(report, ResolutionUnknown, ReasonUnavailable)
	}
	values, err := decode(input)
	if err != nil {
		return lower(report, ResolutionInvariant, ReasonMalformed)
	}
	markDecoded(&report)
	observe(&report, values)
	if unknownDecision(values) {
		return lower(report, ResolutionUnknown, ReasonDecisionUnknown)
	}
	if report.Artifacts[0].ObservedDigest != AssuranceDigest {
		return lower(report, ResolutionInvariant, ReasonDigestMismatch)
	}
	if !validAssurance(values.Assurance) {
		return lower(report, ResolutionInvariant, ReasonAssuranceMismatch)
	}
	if !validSubject(input.SubjectSHA, values) {
		return lower(report, ResolutionInvariant, ReasonSubjectMismatch)
	}
	if !validReference(values) {
		return lower(report, ResolutionInvariant, ReasonReferenceMismatch)
	}
	if report.Summary.RepositoryWrites != 0 || report.Summary.ExternalRepositoryWrites != 0 {
		return lower(report, ResolutionInvariant, ReasonWriteObserved)
	}
	if report.Summary.OfficialMutations != 0 {
		return lower(report, ResolutionInvariant, ReasonMutationObserved)
	}
	if report.Summary.Promotions != 0 {
		return lower(report, ResolutionInvariant, ReasonPromotionObserved)
	}
	if !validParent(values) {
		return lower(report, ResolutionInvariant, ReasonParentMismatch)
	}
	if !validCapability(values) {
		return lower(report, ResolutionInvariant, ReasonCapabilityMismatch)
	}
	report.Decision, report.Resolution = DecisionEligible, ResolutionExact
	report.EnforcementEffect, report.Reason = EffectNone, ReasonEligible
	report.Summary.ProjectedOperating, report.Summary.ProjectedCoverageBPS = 12, 10000
	report.Summary.EligiblePaths = 1
	return finish(report)
}

func baseReport(input Input) Report {
	return Report{Schema: ReportSchema, SubjectSHA: input.SubjectSHA,
		AssuranceSubjectSHA: AssuranceSubject, Decision: DecisionFailClosed,
		Resolution: ResolutionUnknown, EnforcementEffect: EffectBlock, Reason: ReasonUnavailable,
		Artifacts: bindings(input), Transition: Transition{MetricID: MetricID, MetaOperation: MetaOperation,
			FromStatus: "NOT_IMPLEMENTED", FromResolution: "NONE", EligibleStatus: "OPERATING",
			EligibleResolution: ResolutionExact, OfficialStatus: "NOT_IMPLEMENTED", OfficialResolution: "NONE"},
		Summary: Summary{AssuranceDenominator: 12, BeforeOperating: 11, ProjectedOperating: 11,
			OfficialOperating: 11, BeforeCoverageBPS: 9166, ProjectedCoverageBPS: 9166,
			OfficialCoverageBPS: 9166, EvidenceTotal: 7, ParentTotal: 8,
			CapabilityTotal: 10, CapabilityOutcomeTotal: 3, CapabilitySuiteTotal: 15}}
}

func lower(report Report, resolution, reason string) Report {
	report.Resolution, report.Reason = resolution, reason
	if resolution == ResolutionUnknown {
		report.Summary.UnknownPaths = 1
	}
	return finish(report)
}
