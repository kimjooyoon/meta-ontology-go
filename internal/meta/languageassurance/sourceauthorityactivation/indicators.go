package sourceauthorityactivation

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

func buildIndicators(summary Summary, applied int, resolution string) []Indicator {
	return []Indicator{
		indicator("gooo.metric.semantic.source-authority-activation-total.v1", "OUTCOME", "COHERENCE", "apply-source-authority-transition", "transitions", "GREATER_OR_EQUAL", applied, 1, resolution),
		indicator("gooo.metric.evidence.activation-capsule-coverage-bps.v1", "DRIVER", "FOUNDATION", "bind-activation-capsules", "basis_points", "GREATER_OR_EQUAL", summary.CapsuleCoverageBPS, 10000, resolution),
		indicator("gooo.metric.evidence.activation-eligibility-exact.v1", "DRIVER", "FOUNDATION", "consume-merged-eligibility", "receipts", "GREATER_OR_EQUAL", summary.EligibilityExact, 1, resolution),
		indicator("gooo.metric.epistemic.activation-unknown-paths.v1", "GUARDRAIL", "REGRESSION", "preserve-activation-unknown", "paths", "LESS_OR_EQUAL", summary.UnknownPaths, 0, resolution),
		indicator("gooo.metric.effects.activation-repository-writes.v1", "GUARDRAIL", "FOUNDATION", "preserve-activation-read-only", "writes", "LESS_OR_EQUAL", 0, 0, ResolutionExact),
		indicator("gooo.metric.semantic.activation-duplicate-paths.v1", "GUARDRAIL", "COHERENCE", "deny-duplicate-activation", "paths", "LESS_OR_EQUAL", summary.BlockedPaths, 0, resolution),
	}
}

func indicator(id, class, proof, operation, unit, relation string, value, target int, resolution string) Indicator {
	satisfied := value >= target
	if relation == "LESS_OR_EQUAL" {
		satisfied = value <= target
	}
	return Indicator{MetricID: id, Class: class, ProofChoice: proof,
		Producer: "sourceauthorityactivation.Evaluate", Consumer: "language-assurance-gate",
		MetaOperation: operation, Unit: unit, Relation: relation, Resolution: resolution,
		Value: value, Target: target, Satisfied: satisfied}
}
