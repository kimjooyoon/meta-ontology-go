package sourceauthorityupstream

func buildIndicators(receipt Receipt) []Indicator {
	exact := receipt.Observation == ObservationSatisfied && receipt.Resolution == ResolutionExact && receipt.Enforcement == EnforcementAllow
	authority := receipt.Reason != ReasonPolicyInvalid && receipt.Reason != ReasonRequestInvalid && receipt.Reason != ReasonAuthorityScopeMismatch
	digest := exact
	laundering := 0
	if receipt.Observation == ObservationUnknown && receipt.Enforcement != EnforcementBlock {
		laundering = 1
	}
	return []Indicator{
		newIndicator("gooo.metric.evidence.upstream-source-snapshot.coverage-bps.v1", "OUTCOME", "COHERENCE", "observe-upstream-source", "sourceauthorityupstream.Observe", "source-authority-upstream-ci", receipt.Resolution, boolBPS(exact), 10000, "GREATER_OR_EQUAL"),
		newIndicator("gooo.metric.evidence.upstream-authority-binding.coverage-bps.v1", "DRIVER", "FOUNDATION", "bind-upstream-authority", "sourceauthorityupstream.validate", "sourceauthorityupstream.Observe", receipt.Resolution, boolBPS(authority), 10000, "GREATER_OR_EQUAL"),
		newIndicator("gooo.metric.evidence.upstream-digest-binding.coverage-bps.v1", "DRIVER", "FOUNDATION", "bind-upstream-digest", "sourceauthorityupstream.digestBytes", "sourceauthorityupstream.Observe", receipt.Resolution, boolBPS(digest), 10000, "GREATER_OR_EQUAL"),
		newIndicator("gooo.metric.epistemic.upstream-unknown-laundering.v1", "GUARDRAIL", "REGRESSION", "preserve-upstream-unknown", "sourceauthorityupstream.Observe", "language-assurance-gate", ResolutionExact, laundering, 0, "LESS_OR_EQUAL"),
		newIndicator("gooo.metric.effects.upstream-source-writes.v1", "GUARDRAIL", "FOUNDATION", "preserve-upstream-read-only", "sourceauthorityupstream.Observe", "source-authority-upstream-ci", ResolutionExact, receipt.RepositoryWrites, 0, "LESS_OR_EQUAL"),
		newIndicator("gooo.metric.semantic.upstream-promotion-credit.v1", "GUARDRAIL", "COHERENCE", "deny-upstream-promotion-credit", "sourceauthorityupstream.Observe", "language-assurance-readiness", ResolutionExact, receipt.PromotionCreditBPS, 0, "LESS_OR_EQUAL"),
	}
}

func newIndicator(id, class, proof, operation, producer, consumer, resolution string, value, target int, relation string) Indicator {
	satisfied := value >= target
	if relation == "LESS_OR_EQUAL" {
		satisfied = value <= target
	}
	return Indicator{
		MetricID: id, Class: class, ProofChoice: proof, MetaOperation: operation, Producer: producer,
		Consumer: consumer, Resolution: resolution, Unit: indicatorUnit(relation), Value: value,
		Target: target, Relation: relation, Satisfied: satisfied,
	}
}

func boolBPS(value bool) int {
	if value {
		return 10000
	}
	return 0
}

func indicatorUnit(relation string) string {
	if relation == "GREATER_OR_EQUAL" {
		return "basis_points"
	}
	return "paths"
}
