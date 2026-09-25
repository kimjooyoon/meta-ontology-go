package changedsurfacereceipteligibility

func MetaOperations() []MetaOperationBinding {
	return []MetaOperationBinding{
		{ID: "bind-receipt-assurance-baseline", ProofChoice: "FOUNDATION"},
		{ID: "consume-receipt-meta-report", ProofChoice: "FOUNDATION"},
		{ID: "consume-receipt-conformance-suite", ProofChoice: "FOUNDATION"},
		{ID: "evaluate-receipt-eligibility", ProofChoice: "COHERENCE"},
		{ID: "deny-receipt-promotion-side-effects", ProofChoice: "COHERENCE"},
		{ID: "preserve-receipt-eligibility-unknown", ProofChoice: "REGRESSION"},
	}
}

func buildIndicators(summary Summary, resolution string) []Indicator {
	return []Indicator{
		indicator("gooo.metric.semantic.changed-surface-receipt-eligibility-bps.v1", "OUTCOME", "COHERENCE",
			"evaluate-receipt-eligibility", "basis_points", "GREATER_OR_EQUAL", resolution,
			summary.EligiblePaths*10_000, 10_000),
		indicator("gooo.metric.evidence.changed-surface-receipt-capsule-bps.v1", "DRIVER", "FOUNDATION",
			"consume-receipt-conformance-suite", "basis_points", "GREATER_OR_EQUAL", resolution,
			summary.CapsuleCoverageBPS, 10_000),
		indicator("gooo.metric.meta.changed-surface-receipt-operation-bps.v1", "DRIVER", "FOUNDATION",
			"consume-receipt-meta-report", "basis_points", "GREATER_OR_EQUAL", resolution,
			summary.MetaOperationCoverageBPS, 10_000),
		indicator("gooo.metric.epistemic.changed-surface-receipt-eligibility-unknown.v1", "GUARDRAIL", "REGRESSION",
			"preserve-receipt-eligibility-unknown", "paths", "LESS_OR_EQUAL", ResolutionExact,
			summary.UnknownPaths, 0),
		indicator("gooo.metric.effects.changed-surface-receipt-eligibility-writes.v1", "GUARDRAIL", "FOUNDATION",
			"bind-receipt-assurance-baseline", "writes", "LESS_OR_EQUAL", ResolutionExact, 0, 0),
		indicator("gooo.metric.semantic.changed-surface-receipt-eligibility-applied.v1", "GUARDRAIL", "COHERENCE",
			"deny-receipt-promotion-side-effects", "transitions", "LESS_OR_EQUAL", ResolutionExact, 0, 0),
	}
}

func indicator(id, class, proof, operation, unit, relation, resolution string, value, target int) Indicator {
	satisfied := value <= target
	if relation == "GREATER_OR_EQUAL" {
		satisfied = value >= target
	}
	return Indicator{MetricID: id, Class: class, ProofChoice: proof,
		Producer: "changedsurfacereceipteligibility.Evaluate", Consumer: "language-assurance-promotion-gate",
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
