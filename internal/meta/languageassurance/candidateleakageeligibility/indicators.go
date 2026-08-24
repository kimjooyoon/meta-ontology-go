package candidateleakageeligibility

func buildIndicators(summary Summary, resolution string) []Indicator {
	return []Indicator{
		indicator("gooo.metric.semantic.candidate-leakage-eligibility-bps.v1", "OUTCOME", "COHERENCE",
			"evaluate-candidate-leakage-eligibility", "basis_points", "GREATER_OR_EQUAL", resolution,
			summary.EligiblePaths*10_000, 10_000),
		indicator("gooo.metric.evidence.candidate-assurance-binding-bps.v1", "DRIVER", "FOUNDATION",
			"bind-candidate-eligibility-baseline", "basis_points", "GREATER_OR_EQUAL", resolution,
			summary.CapsuleCoverageBPS, 10_000),
		indicator("gooo.metric.evidence.candidate-shadow-binding-bps.v1", "DRIVER", "FOUNDATION",
			"consume-merged-candidate-shadow", "basis_points", "GREATER_OR_EQUAL", resolution,
			summary.CapsuleCoverageBPS, 10_000),
		indicator("gooo.metric.epistemic.candidate-eligibility-unknown.v1", "GUARDRAIL", "REGRESSION",
			"preserve-eligibility-unknown", "paths", "LESS_OR_EQUAL", ResolutionExact, summary.UnknownPaths, 0),
		indicator("gooo.metric.effects.candidate-eligibility-writes.v1", "GUARDRAIL", "FOUNDATION",
			"preserve-eligibility-read-only", "writes", "LESS_OR_EQUAL", ResolutionExact, 0, 0),
		indicator("gooo.metric.semantic.candidate-eligibility-applied.v1", "GUARDRAIL", "COHERENCE",
			"deny-eligibility-side-effects", "transitions", "LESS_OR_EQUAL", ResolutionExact, 0, 0),
	}
}

func indicator(id, class, proof, operation, unit, relation, resolution string, value, target int) Indicator {
	satisfied := value <= target
	if relation == "GREATER_OR_EQUAL" {
		satisfied = value >= target
	}
	return Indicator{MetricID: id, Class: class, ProofChoice: proof,
		Producer: "candidateleakageeligibility.Evaluate", Consumer: "language-assurance-promotion-gate",
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
