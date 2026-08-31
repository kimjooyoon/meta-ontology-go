package rollbackintegrityeligibility

func MetaOperations() []MetaOperationBinding {
	return []MetaOperationBinding{
		{ID: "bind-rollback-assurance-baseline", ProofChoice: "FOUNDATION"},
		{ID: "consume-rollback-shadow-replay-a", ProofChoice: "FOUNDATION"},
		{ID: "consume-rollback-shadow-replay-b", ProofChoice: "FOUNDATION"},
		{ID: "compare-rollback-shadow-replays", ProofChoice: "COHERENCE"},
		{ID: "deny-rollback-promotion-side-effects", ProofChoice: "COHERENCE"},
		{ID: "evaluate-rollback-eligibility", ProofChoice: "REGRESSION"},
	}
}

func buildIndicators(summary Summary, resolution string) []Indicator {
	return []Indicator{
		indicator("gooo.metric.operation.rollback-integrity.eligibility-bps.v1", "OUTCOME", "COHERENCE",
			"compare-rollback-shadow-replays", "basis_points", "GREATER_OR_EQUAL", resolution,
			summary.EligiblePaths*10_000, 10_000),
		indicator("gooo.metric.evidence.rollback-integrity-eligibility-capsule-bps.v1", "DRIVER", "FOUNDATION",
			"consume-rollback-shadow-replay-b", "basis_points", "GREATER_OR_EQUAL", resolution,
			summary.CapsuleCoverageBPS, 10_000),
		indicator("gooo.metric.meta.rollback-integrity-eligibility-operation-bps.v1", "DRIVER", "FOUNDATION",
			"consume-rollback-shadow-replay-a", "basis_points", "GREATER_OR_EQUAL", resolution,
			summary.MetaOperationCoverageBPS, 10_000),
		indicator("gooo.metric.epistemic.rollback-integrity-eligibility-unknown.v1", "GUARDRAIL", "REGRESSION",
			"evaluate-rollback-eligibility", "paths", "LESS_OR_EQUAL", ResolutionExact,
			summary.UnknownPaths, 0),
		indicator("gooo.metric.effects.rollback-integrity-eligibility-writes.v1", "GUARDRAIL", "FOUNDATION",
			"bind-rollback-assurance-baseline", "writes", "LESS_OR_EQUAL", ResolutionExact, 0, 0),
		indicator("gooo.metric.operation.rollback-integrity-eligibility-applied.v1", "GUARDRAIL", "COHERENCE",
			"deny-rollback-promotion-side-effects", "transitions", "LESS_OR_EQUAL", ResolutionExact, 0, 0),
	}
}

func indicator(id, class, proof, operation, unit, relation, resolution string, value, target int) Indicator {
	satisfied := value <= target
	if relation == "GREATER_OR_EQUAL" {
		satisfied = value >= target
	}
	return Indicator{MetricID: id, Class: class, ProofChoice: proof,
		Producer: "rollbackintegrityeligibility.Evaluate", Consumer: "language-assurance-promotion-gate",
		MetaOperation: operation, Unit: unit, Relation: relation, Resolution: resolution,
		Value: value, Target: target, Satisfied: satisfied}
}

type Indicator struct {
	MetricID      string `json:"metric_id"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	Unit          string `json:"unit"`
	Relation      string `json:"relation"`
	Resolution    string `json:"resolution"`
	Value         int    `json:"value"`
	Target        int    `json:"target"`
	Satisfied     bool   `json:"satisfied"`
}
