package candidateleakage

func buildIndicators(summary Summary, resolution string) []Indicator {
	return []Indicator{
		indicator("gooo.metric.semantic.candidate-leakage-paths.v1", "OUTCOME", "COHERENCE",
			"detect-candidate-leakage", "paths", "LESS_OR_EQUAL", resolution, summary.LeakagePaths, 0),
		indicator("gooo.metric.evidence.candidate-boundary-binding-bps.v1", "DRIVER", "FOUNDATION",
			"observe-candidate-envelope", "basis_points", "GREATER_OR_EQUAL", resolution, summary.BoundaryBindingBPS, 10_000),
		indicator("gooo.metric.evidence.promotion-authority-binding-bps.v1", "DRIVER", "FOUNDATION",
			"bind-promotion-authority", "basis_points", "GREATER_OR_EQUAL", resolution, summary.PromotionAuthorityBPS, 10_000),
		indicator("gooo.metric.epistemic.candidate-unknown-laundering.v1", "GUARDRAIL", "REGRESSION",
			"preserve-candidate-unknown", "paths", "LESS_OR_EQUAL", ResolutionExact, summary.UnknownPaths, 0),
		indicator("gooo.metric.effects.candidate-leakage-writes.v1", "GUARDRAIL", "FOUNDATION",
			"preserve-candidate-read-only", "writes", "LESS_OR_EQUAL", ResolutionExact, summary.RepositoryWrites, 0),
		indicator("gooo.metric.semantic.candidate-shadow-promotion-credit.v1", "GUARDRAIL", "COHERENCE",
			"deny-shadow-promotion-credit", "basis_points", "LESS_OR_EQUAL", ResolutionExact, summary.PromotionCreditBPS, 0),
	}
}

func indicator(id, class, proof, operation, unit, relation, resolution string, value, target int) Indicator {
	satisfied := value <= target
	if relation == "GREATER_OR_EQUAL" {
		satisfied = value >= target
	}
	return Indicator{
		MetricID: id, Class: class, ProofChoice: proof,
		Producer: "candidateleakage.Evaluate", Consumer: "candidate-leakage-shadow-gate",
		MetaOperation: operation, Unit: unit, Relation: relation, Resolution: resolution,
		Value: value, Target: target, Satisfied: satisfied,
	}
}
