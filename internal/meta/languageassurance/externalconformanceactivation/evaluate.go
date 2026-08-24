package externalconformanceactivation

import "fmt"

func Evaluate(input Input) Receipt {
	state := validateEvidence(input)
	applied := state.Reason == ""
	decision, effect, reason, after, transitioned, blocked := DecisionFailClosed, EffectBlock, state.Reason, 11, 0, 1
	if applied {
		decision, effect, reason, after, transitioned, blocked = DecisionApplied, EffectApply, ReasonApplied, 12, 1, 0
	}
	unknown := 0
	if state.Resolution == ResolutionUnknown {
		unknown = 1
	}
	s := state.Eligibility.Summary
	semanticsExact := state.AssuranceExact + state.EligibilityExact + state.MergeExact
	summary := Summary{DenominatorTotal: 12, BeforeOperating: 11, AfterOperating: after,
		BeforeCoverageBPS: 9166, AfterCoverageBPS: after * 10000 / 12,
		CapsulesTotal: 3, CapsulesExact: state.RawExact, CapsuleCoverageBPS: state.RawExact * 10000 / 3,
		PredecessorSemanticsBPS: semanticsExact * 10000 / 3,
		EligibilityIndicatorsTotal: s.IndicatorTotal, EligibilityIndicatorsSatisfied: s.IndicatorCompleted,
		ParentCompleted: s.ParentCompleted, ParentTotal: s.ParentTotal, ParentKnownFailures: s.ParentKnownFailures,
		SelectedCompleted: s.CapabilityCompleted, SelectedTotal: s.CapabilityTotal, ExternalExecutions: s.ExternalExecutions,
		MergeRelationsTotal: 1, MergeRelationsSatisfied: state.MergeExact, UnknownPaths: unknown, BlockedPaths: blocked}
	receipt := Receipt{Schema: Schema, SubjectSHA: input.SubjectSHA, PredecessorSHA: PredecessorSHA,
		EligibilitySubjectSHA: EligibilitySubjectSHA, Decision: decision, Resolution: state.Resolution,
		EnforcementEffect: effect, Reason: reason, DenominatorID: DenominatorID,
		DenominatorDigest: digestValue(Denominator()), EligibilityReportDigest: EligibilityReportHash,
		Artifacts: artifactBindings(input), Transition: Transition{MetricID: MetricID, MetaOperation: MetaOperation,
			FromStatus: "NOT_IMPLEMENTED", FromResolution: "NONE", ToStatus: "OPERATING", ToResolution: ResolutionExact},
		Summary: summary, Indicators: buildIndicators(summary, transitioned, state.Resolution),
		MetaOperations: activationMetaOperations(), RepositoryWrites: 0, TransitionApplied: transitioned}
	seal(&receipt)
	return receipt
}

func OperatingOperation() (string, string, error) {
	receipt := Evaluate(EmbeddedInput(PredecessorSHA))
	if receipt.Decision != DecisionApplied || receipt.Resolution != ResolutionExact ||
		receipt.EnforcementEffect != EffectApply || receipt.TransitionApplied != 1 {
		return "", "", fmt.Errorf("%s: %s", receipt.Reason, receipt.Resolution)
	}
	return receipt.Transition.MetricID, receipt.Transition.MetaOperation, nil
}
