package languageassurance

import "slices"

func buildIndicators(summary Summary) []Indicator {
	coverage := summary.ImplementationCoverageBPS
	evidence := summary.EvidenceCoverageBPS
	return []Indicator{
		indicator("gooo.metric.assurance.implementation-coverage-bps.v1", ClassOutcome, ProofCoherence, "freeze-assurance-denominator", &coverage, 10000, "basis_points", RelationGreaterOrEqual),
		indicator("gooo.metric.assurance.transaction-evidence-coverage-bps.v1", ClassDriver, ProofFoundation, "observe-transaction-evidence", &evidence, 10000, "basis_points", RelationGreaterOrEqual),
		indicator(MetricSnapshotBinding, ClassDriver, ProofFoundation, "bind-exact-snapshot", summary.ExactSnapshotBindingBPS, 10000, "basis_points", RelationGreaterOrEqual),
		indicator(MetricSelfMinting, ClassGuardrail, ProofFoundation, "detect-self-minting-paths", summary.SelfMintingPaths, 0, "paths", RelationLessOrEqual),
		indicator(MetricRoleConflict, ClassGuardrail, ProofCoherence, "detect-role-conflict-paths", summary.RoleConflictPaths, 0, "paths", RelationLessOrEqual),
		indicator(MetricUnknownLaundering, ClassGuardrail, ProofRegression, "detect-unknown-laundering", summary.UnknownLaunderingPaths, 0, "paths", RelationLessOrEqual),
	}
}

func indicator(metricID string, class IndicatorClass, proof ProofChoice, operation string, value *int, target int, unit string, relation Relation) Indicator {
	resolution := ResolutionExact
	if value == nil {
		resolution = ResolutionUnknown
	}
	return Indicator{MetricID: metricID, Class: class, ProofChoice: proof, Producer: Producer, Consumer: Consumer, MetaOperation: operation, Value: value, Target: target, Unit: unit, Relation: relation, Resolution: resolution, Satisfied: satisfies(value, target, relation)}
}

func decide(summary Summary) (string, string, Resolution) {
	if summary.UnresolvedIndicators > 0 || summary.EvidenceGroupsObserved != summary.EvidenceGroupsTotal {
		return CandidateFailClosed, ReasonEvidenceUnknown, ResolutionUnknown
	}
	if summary.UnknownTopDecisions != nil && *summary.UnknownTopDecisions > 0 {
		return CandidateFailClosed, ReasonTopDecisionUnknown, ResolutionInvariantOnly
	}
	if summary.ExactSnapshotBindingBPS != nil && *summary.ExactSnapshotBindingBPS < 10000 {
		return CandidateBlock, ReasonSnapshotMismatch, ResolutionExact
	}
	if summary.ViolatedGuardrails > 0 {
		return CandidateBlock, ReasonGovernanceViolation, ResolutionExact
	}
	return CandidateAllowLimited, ReasonBoundaryClear, ResolutionExact
}

func isLaunderingOutput(output Decision) bool {
	return slices.Contains(launderingOutputs, output)
}

func unresolved(values ...*int) int {
	count := 0
	for _, value := range values {
		if value == nil {
			count++
		}
	}
	return count
}

func positive(values ...*int) int {
	count := 0
	for _, value := range values {
		if value != nil && *value > 0 {
			count++
		}
	}
	return count
}

func satisfies(value *int, target int, relation Relation) bool {
	if value == nil {
		return false
	}
	if relation == RelationGreaterOrEqual {
		return *value >= target
	}
	return *value <= target
}
