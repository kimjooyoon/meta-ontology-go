package sourceauthorityactivation

import "fmt"

func Evaluate(input Input) Receipt {
	_, exact, reason, resolution := validateEvidence(input)
	applied := reason == ""
	decision, transitionApplied, blocked, eligibilityExact := DecisionFailClosed, 0, 1, 0
	after := 6
	if applied {
		decision, reason, resolution = DecisionApplied, ReasonApplied, ResolutionExact
		transitionApplied, blocked, eligibilityExact, after = 1, 0, 1, 7
	}
	unknown := 0
	if resolution == ResolutionUnknown {
		unknown = 1
	}
	summary := Summary{DenominatorTotal: 12, BeforeOperating: 6, AfterOperating: after,
		BeforeCoverageBPS: 5000, AfterCoverageBPS: after * 10000 / 12,
		CapsulesTotal: 3, CapsulesExact: exact, CapsuleCoverageBPS: exact * 10000 / 3,
		EligibilityExact: eligibilityExact, UnknownPaths: unknown, BlockedPaths: blocked}
	receipt := Receipt{Schema: Schema, SubjectSHA: input.SubjectSHA, PredecessorSHA: PredecessorSHA,
		Decision: decision, Resolution: resolution, Reason: reason,
		DenominatorID: DenominatorID, DenominatorDigest: digestValue(Denominator()),
		EligibilityReportDigest: EligibilityReportHash, Artifacts: artifactBindings(input),
		Transition: Transition{MetricID: MetricID, MetaOperation: MetaOperation,
			FromStatus: "NOT_IMPLEMENTED", FromResolution: "NONE", ToStatus: "OPERATING", ToResolution: ResolutionExact},
		Summary: summary, RepositoryWrites: 0, TransitionApplied: transitionApplied,
		Indicators: buildIndicators(summary, transitionApplied, resolution)}
	seal(&receipt)
	return receipt
}

func OperatingOperation() (string, string, error) {
	receipt := Evaluate(EmbeddedInput(PredecessorSHA))
	if receipt.Decision != DecisionApplied || receipt.Resolution != ResolutionExact || receipt.TransitionApplied != 1 {
		return "", "", fmt.Errorf("%s: %s", receipt.Reason, receipt.Resolution)
	}
	return receipt.Transition.MetricID, receipt.Transition.MetaOperation, nil
}
