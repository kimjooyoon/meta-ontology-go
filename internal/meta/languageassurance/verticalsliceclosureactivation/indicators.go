package verticalsliceclosureactivation

func buildIndicators(summary Summary, applied int, resolution string) []Indicator {
	return []Indicator{
		indicator("gooo.metric.capability.vertical-slice-activation-total.v1", "OUTCOME", "COHERENCE", "apply-vertical-slice-transition", "transitions", "GREATER_OR_EQUAL", applied, 1, resolution),
		indicator("gooo.metric.evidence.vertical-slice-activation-capsule-bps.v1", "DRIVER", "FOUNDATION", "bind-vertical-slice-activation-capsules", "basis_points", "GREATER_OR_EQUAL", summary.CapsuleCoverageBPS, 10000, resolution),
		indicator("gooo.metric.evidence.vertical-slice-predecessor-semantics-bps.v1", "DRIVER", "FOUNDATION", "consume-vertical-slice-predecessor-semantics", "basis_points", "GREATER_OR_EQUAL", summary.PredecessorSemanticsBPS, 10000, resolution),
		indicator("gooo.metric.epistemic.vertical-slice-activation-unknown.v1", "GUARDRAIL", "REGRESSION", "preserve-vertical-slice-activation-unknown", "paths", "LESS_OR_EQUAL", summary.UnknownPaths, 0, resolution),
		indicator("gooo.metric.effects.vertical-slice-activation-writes.v1", "GUARDRAIL", "FOUNDATION", "preserve-vertical-slice-activation-read-only", "writes", "LESS_OR_EQUAL", 0, 0, ResolutionExact),
		indicator("gooo.metric.capability.vertical-slice-activation-blocked.v1", "GUARDRAIL", "COHERENCE", "deny-vertical-slice-activation-without-proof", "paths", "LESS_OR_EQUAL", summary.BlockedPaths, 0, resolution),
	}
}

func activationMetaOperations() []MetaOperationBinding {
	return []MetaOperationBinding{
		{ID: "bind-vertical-slice-activation-capsules", ProofChoice: "FOUNDATION"},
		{ID: "consume-vertical-slice-predecessor-semantics", ProofChoice: "FOUNDATION"},
		{ID: "preserve-vertical-slice-activation-read-only", ProofChoice: "FOUNDATION"},
		{ID: "apply-vertical-slice-transition", ProofChoice: "COHERENCE"},
		{ID: "deny-vertical-slice-activation-without-proof", ProofChoice: "COHERENCE"},
		{ID: "preserve-vertical-slice-activation-unknown", ProofChoice: "REGRESSION"},
	}
}

func indicator(id, class, proof, operation, unit, relation string, value, target int, resolution string) Indicator {
	satisfied := value >= target
	if relation == "LESS_OR_EQUAL" {
		satisfied = value <= target
	}
	return Indicator{MetricID: id, Class: class, ProofChoice: proof,
		Producer: "verticalsliceclosureactivation.Evaluate", Consumer: "language-assurance-gate",
		MetaOperation: operation, Unit: unit, Relation: relation, Resolution: resolution,
		Value: value, Target: target, Satisfied: satisfied}
}
