package sourceauthoritypromotion

type Indicator struct {
	MetricID, Class, ProofChoice, Producer, Consumer string
	MetaOperation, Unit, Relation, Resolution        string
	Value, Target                                   int
	Satisfied                                       bool
}

func buildIndicators(baselineOK, evidenceOK, eligible bool) []Indicator {
	return []Indicator{
		indicator("gooo.metric.semantic.source-authority-promotion-eligibility-bps.v1", "OUTCOME", "COHERENCE", "evaluate-promotion-eligibility", eligible, true),
		indicator("gooo.metric.evidence.assurance-denominator-binding-bps.v1", "DRIVER", "FOUNDATION", "bind-assurance-denominator", baselineOK, true),
		indicator("gooo.metric.evidence.upstream-conformance-binding-bps.v1", "DRIVER", "FOUNDATION", "bind-upstream-conformance", evidenceOK, true),
		guardrail("gooo.metric.epistemic.promotion-unknown-laundering.v1", "REGRESSION", "preserve-promotion-unknown"),
		guardrail("gooo.metric.effects.promotion-eligibility-writes.v1", "FOUNDATION", "preserve-eligibility-read-only"),
		guardrail("gooo.metric.semantic.promotion-applied-paths.v1", "COHERENCE", "deny-eligibility-side-effect"),
	}
}

func indicator(id, class, proof, operation string, satisfied, basisPoints bool) Indicator {
	value, resolution := 0, ResolutionInvariantOnly
	if satisfied {
		value, resolution = 10000, ResolutionExact
	}
	unit := "basis_points"
	if !basisPoints { unit = "paths" }
	return Indicator{MetricID: id, Class: class, ProofChoice: proof,
		Producer: "sourceauthoritypromotion.Evaluate", Consumer: "language-assurance-promotion-gate",
		MetaOperation: operation, Unit: unit, Relation: "GREATER_OR_EQUAL",
		Resolution: resolution, Value: value, Target: 10000, Satisfied: satisfied}
}

func guardrail(id, proof, operation string) Indicator {
	return Indicator{MetricID: id, Class: "GUARDRAIL", ProofChoice: proof,
		Producer: "sourceauthoritypromotion.Evaluate", Consumer: "language-assurance-promotion-gate",
		MetaOperation: operation, Unit: "paths", Relation: "LESS_OR_EQUAL",
		Resolution: ResolutionExact, Value: 0, Target: 0, Satisfied: true}
}
