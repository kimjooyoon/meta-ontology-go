package externalconformanceactivation

func buildIndicators(summary Summary, applied int, resolution string) []Indicator {
	return []Indicator{
		indicator("gooo.metric.capability.external-activation-total.v1", "OUTCOME", "COHERENCE", "apply-external-conformance-transition", "transitions", "GREATER_OR_EQUAL", applied, 1, resolution),
		indicator("gooo.metric.assurance.external-activation-operating.v1", "OUTCOME", "COHERENCE", "reconstruct-external-assurance", "obligations", "GREATER_OR_EQUAL", summary.AfterOperating, 12, resolution),
		indicator("gooo.metric.evidence.external-activation-capsule-bps.v1", "DRIVER", "FOUNDATION", "bind-external-activation-capsules", "basis_points", "GREATER_OR_EQUAL", summary.CapsuleCoverageBPS, 10000, resolution),
		indicator("gooo.metric.evidence.external-predecessor-semantics-bps.v1", "DRIVER", "FOUNDATION", "consume-external-predecessor-semantics", "basis_points", "GREATER_OR_EQUAL", summary.PredecessorSemanticsBPS, 10000, resolution),
		indicator("gooo.metric.evidence.external-eligibility-indicators.v1", "DRIVER", "FOUNDATION", "consume-external-eligibility-indicators", "indicators", "GREATER_OR_EQUAL", summary.EligibilityIndicatorsSatisfied, 18, resolution),
		indicator("gooo.metric.evidence.external-merge-relation.v1", "DRIVER", "FOUNDATION", "bind-external-eligibility-merge", "relations", "GREATER_OR_EQUAL", summary.MergeRelationsSatisfied, 1, resolution),
		indicator("gooo.metric.epistemic.external-activation-unknown.v1", "GUARDRAIL", "REGRESSION", "preserve-external-activation-unknown", "paths", "LESS_OR_EQUAL", summary.UnknownPaths, 0, resolution),
		indicator("gooo.metric.effects.external-activation-writes.v1", "GUARDRAIL", "REGRESSION", "preserve-external-activation-read-only", "writes", "LESS_OR_EQUAL", 0, 0, ResolutionExact),
		indicator("gooo.metric.guardrail.external-parent-failures.v1", "GUARDRAIL", "REGRESSION", "preserve-external-parent-failures", "failures", "EQUAL", summary.ParentKnownFailures, 2, resolution),
		indicator("gooo.metric.guardrail.external-single-transition.v1", "GUARDRAIL", "REGRESSION", "bound-external-activation-transition", "transitions", "LESS_OR_EQUAL", applied, 1, resolution),
	}
}

func activationMetaOperations() []MetaOperationBinding {
	return []MetaOperationBinding{
		{"bind-external-activation-capsules", "FOUNDATION"},
		{"consume-external-predecessor-semantics", "FOUNDATION"},
		{"consume-external-eligibility-indicators", "FOUNDATION"},
		{"bind-external-eligibility-merge", "FOUNDATION"},
		{"apply-external-conformance-transition", "COHERENCE"},
		{"reconstruct-external-assurance", "COHERENCE"},
		{"preserve-external-activation-unknown", "REGRESSION"},
		{"preserve-external-activation-read-only", "REGRESSION"},
		{"preserve-external-parent-failures", "REGRESSION"},
		{"bound-external-activation-transition", "REGRESSION"},
	}
}

func indicator(id, class, proof, operation, unit, relation string, value, target int, resolution string) Indicator {
	satisfied := value >= target
	if relation == "LESS_OR_EQUAL" {
		satisfied = value <= target
	} else if relation == "EQUAL" {
		satisfied = value == target
	}
	return Indicator{MetricID: id, Class: class, ProofChoice: proof,
		Producer: "externalconformanceactivation.Evaluate", Consumer: "language-assurance-gate",
		MetaOperation: operation, Unit: unit, Relation: relation, Resolution: resolution,
		Value: value, Target: target, Satisfied: satisfied}
}
