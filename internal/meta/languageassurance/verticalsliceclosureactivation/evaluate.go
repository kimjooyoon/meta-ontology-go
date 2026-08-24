package verticalsliceclosureactivation

import "fmt"

func Evaluate(input Input) Receipt {
	report, rawExact, assuranceExact, eligibilityExact, reason, resolution := validateEvidence(input)
	applied := reason == ""
	decision, effect, transitionApplied, blocked, after := DecisionFailClosed, EffectBlock, 0, 1, 10
	if applied {
		decision, effect, reason, resolution = DecisionApplied, EffectApply, ReasonApplied, ResolutionExact
		transitionApplied, blocked, after = 1, 0, 11
	}
	unknown := 0
	if resolution == ResolutionUnknown { unknown = 1 }
	boundariesTotal, boundariesSatisfied, linksTotal, linksSatisfied := 0, 0, 0, 0
	indicatorTotal, indicatorSatisfied, operationObserved := 0, 0, 0
	if eligibilityExact == 1 {
		s := report.Summary
		boundariesTotal, boundariesSatisfied = s.BoundariesTotal, s.BoundariesSatisfied
		linksTotal, linksSatisfied = s.LinksTotal, s.LinksSatisfied
		indicatorTotal, indicatorSatisfied = len(report.Indicators), countSatisfied(report.Indicators)
		operationObserved = len(report.MetaOperations)
	}
	summary := Summary{DenominatorTotal: 12, BeforeOperating: 10, AfterOperating: after,
		BeforeCoverageBPS: 8333, AfterCoverageBPS: after * 10000 / 12,
		CapsulesTotal: 2, CapsulesExact: rawExact, CapsuleCoverageBPS: rawExact * 5000,
		PredecessorSemanticsBPS: (assuranceExact + eligibilityExact) * 5000,
		BoundariesTotal: boundariesTotal, BoundariesSatisfied: boundariesSatisfied,
		LinksTotal: linksTotal, LinksSatisfied: linksSatisfied,
		EligibilityIndicatorsTotal: indicatorTotal, EligibilityIndicatorsSatisfied: indicatorSatisfied,
		MetaOperationsRequired: 6, MetaOperationsObserved: operationObserved,
		UnknownPaths: unknown, BlockedPaths: blocked}
	receipt := Receipt{Schema: Schema, SubjectSHA: input.SubjectSHA, PredecessorSHA: PredecessorSHA,
		Decision: decision, Resolution: resolution, EnforcementEffect: effect, Reason: reason,
		DenominatorID: DenominatorID, DenominatorDigest: digestValue(Denominator()),
		EligibilityReportDigest: EligibilityReportHash, Artifacts: artifactBindings(input),
		Transition: Transition{MetricID: MetricID, MetaOperation: MetaOperation, FromStatus: "NOT_IMPLEMENTED", FromResolution: "NONE", ToStatus: "OPERATING", ToResolution: ResolutionExact},
		Summary: summary, Indicators: buildIndicators(summary, transitionApplied, resolution), MetaOperations: activationMetaOperations(),
		RepositoryWrites: 0, TransitionApplied: transitionApplied}
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
