package rollbackintegrityactivation

func buildIndicators(summary Summary, applied int, resolution string) []Indicator {
	return []Indicator{
		indicator("gooo.metric.operation.rollback-integrity-activation-total.v1", "OUTCOME", "COHERENCE", "apply-rollback-integrity-transition", "transitions", "GREATER_OR_EQUAL", applied, 1, resolution),
		indicator("gooo.metric.evidence.rollback-activation-capsule-bps.v1", "DRIVER", "FOUNDATION", "bind-rollback-activation-capsules", "basis_points", "GREATER_OR_EQUAL", summary.CapsuleCoverageBPS, 10000, resolution),
		indicator("gooo.metric.evidence.rollback-predecessor-semantics-bps.v1", "DRIVER", "FOUNDATION", "consume-rollback-predecessor-semantics", "basis_points", "GREATER_OR_EQUAL", summary.PredecessorSemanticsBPS, 10000, resolution),
		indicator("gooo.metric.epistemic.rollback-activation-unknown.v1", "GUARDRAIL", "REGRESSION", "preserve-rollback-activation-unknown", "paths", "LESS_OR_EQUAL", summary.UnknownPaths, 0, resolution),
		indicator("gooo.metric.effects.rollback-activation-writes.v1", "GUARDRAIL", "FOUNDATION", "preserve-rollback-activation-read-only", "writes", "LESS_OR_EQUAL", 0, 0, ResolutionExact),
		indicator("gooo.metric.operation.rollback-activation-blocked.v1", "GUARDRAIL", "COHERENCE", "deny-rollback-activation-without-proof", "paths", "LESS_OR_EQUAL", summary.BlockedPaths, 0, resolution),
	}
}

func activationMetaOperations() []MetaOperationBinding {
	return []MetaOperationBinding{
		{ID: "bind-rollback-activation-capsules", ProofChoice: "FOUNDATION"},
		{ID: "consume-rollback-predecessor-semantics", ProofChoice: "FOUNDATION"},
		{ID: "preserve-rollback-activation-read-only", ProofChoice: "FOUNDATION"},
		{ID: "apply-rollback-integrity-transition", ProofChoice: "COHERENCE"},
		{ID: "deny-rollback-activation-without-proof", ProofChoice: "COHERENCE"},
		{ID: "preserve-rollback-activation-unknown", ProofChoice: "REGRESSION"},
	}
}

func indicator(id, class, proof, operation, unit, relation string, value, target int, resolution string) Indicator {
	satisfied := value >= target
	if relation == "LESS_OR_EQUAL" {
		satisfied = value <= target
	}
	return Indicator{MetricID: id, Class: class, ProofChoice: proof,
		Producer: "rollbackintegrityactivation.Evaluate", Consumer: "language-assurance-gate",
		MetaOperation: operation, Unit: unit, Relation: relation, Resolution: resolution,
		Value: value, Target: target, Satisfied: satisfied}
}

func validEligibilityIndicators(indicators []eligibilityIndicator) bool {
	classes, proofs := map[string]int{}, map[string]int{}
	for _, item := range indicators {
		if !item.Satisfied || item.Resolution != ResolutionExact || item.Producer != "rollbackintegrityeligibility.Evaluate" || item.Consumer != "language-assurance-promotion-gate" || item.MetaOperation == "" {
			return false
		}
		classes[item.Class]++
		proofs[item.ProofChoice]++
	}
	return len(indicators) == 6 && classes["OUTCOME"] == 1 && classes["DRIVER"] == 2 && classes["GUARDRAIL"] == 3 &&
		proofs["FOUNDATION"] == 3 && proofs["COHERENCE"] == 2 && proofs["REGRESSION"] == 1
}

func validEligibilityMetaOperations(operations []eligibilityOperation) bool {
	proofs := map[string]int{}
	for _, operation := range operations {
		if operation.ID == "" {
			return false
		}
		proofs[operation.ProofChoice]++
	}
	return len(operations) == 6 && proofs["FOUNDATION"] == 3 && proofs["COHERENCE"] == 2 && proofs["REGRESSION"] == 1
}
