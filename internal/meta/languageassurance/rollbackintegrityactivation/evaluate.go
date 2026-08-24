package rollbackintegrityactivation

import "fmt"

func Evaluate(input Input) Receipt {
	report, rawExact, assuranceExact, eligibilityExact, reason, resolution := validateEvidence(input)
	applied := reason == ""
	decision, effect, transitionApplied, blocked, after := DecisionFailClosed, EffectBlock, 0, 1, 9
	if applied {
		decision, effect, reason, resolution = DecisionApplied, EffectApply, ReasonApplied, ResolutionExact
		transitionApplied, blocked, after = 1, 0, 10
	}
	unknown := 0
	if resolution == ResolutionUnknown {
		unknown = 1
	}
	semanticExact := assuranceExact + eligibilityExact
	observedOperations := 0
	shadowTotal, shadowPassed, replayTotal, replayExact := 0, 0, 0, 0
	if eligibilityExact == 1 {
		observedOperations = 6
		shadowTotal, shadowPassed = report.Summary.ShadowCasesTotal, report.Summary.ShadowCasesPassed
		replayTotal, replayExact = report.Summary.ShadowReplaysTotal, report.Summary.ShadowReplaysExact
	}
	summary := Summary{
		DenominatorTotal: 12, BeforeOperating: 9, AfterOperating: after,
		BeforeCoverageBPS: 7500, AfterCoverageBPS: after * 10000 / 12,
		CapsulesTotal: 2, CapsulesExact: rawExact, CapsuleCoverageBPS: rawExact * 10000 / 2,
		PredecessorSemanticsBPS: semanticExact * 10000 / 2,
		ShadowCasesTotal:        shadowTotal, ShadowCasesPassed: shadowPassed,
		ShadowReplaysTotal: replayTotal, ShadowReplaysExact: replayExact,
		MetaOperationsRequired: 6, MetaOperationsObserved: observedOperations,
		MetaOperationCoverageBPS: observedOperations * 10000 / 6,
		UnknownPaths:             unknown, BlockedPaths: blocked,
	}
	receipt := Receipt{
		Schema: Schema, SubjectSHA: input.SubjectSHA, PredecessorSHA: PredecessorSHA,
		Decision: decision, Resolution: resolution, EnforcementEffect: effect, Reason: reason,
		DenominatorID: DenominatorID, DenominatorDigest: digestValue(Denominator()),
		EligibilityReportDigest: EligibilityReportHash, Artifacts: artifactBindings(input),
		Transition: Transition{MetricID: MetricID, MetaOperation: MetaOperation,
			FromStatus: "NOT_IMPLEMENTED", FromResolution: "NONE", ToStatus: "OPERATING", ToResolution: ResolutionExact},
		Summary: summary, Indicators: buildIndicators(summary, transitionApplied, resolution),
		MetaOperations: activationMetaOperations(), RepositoryWrites: 0, TransitionApplied: transitionApplied,
	}
	seal(&receipt)
	return receipt
}

func OperatingOperation() (string, string, error) {
	receipt := Evaluate(EmbeddedInput(PredecessorSHA))
	if receipt.Decision != DecisionApplied || receipt.Resolution != ResolutionExact || receipt.EnforcementEffect != EffectApply || receipt.TransitionApplied != 1 {
		return "", "", fmt.Errorf("%s: %s", receipt.Reason, receipt.Resolution)
	}
	return receipt.Transition.MetricID, receipt.Transition.MetaOperation, nil
}
