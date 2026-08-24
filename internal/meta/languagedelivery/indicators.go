package languagedelivery

func MetricIDs() []string {
	return []string{
		"gooo.metric.language.delivery-completed.v1",
		"gooo.metric.language.delivery-readiness-bps.v1",
		"gooo.metric.language.delivery-meta-bindings.v1",
		"gooo.metric.language.delivery-source-receipts.v1",
		"gooo.metric.language.delivery-reader-projections.v1",
		"gooo.metric.language.delivery-unknown.guardrail.v1",
		"gooo.metric.language.delivery-effects.guardrail.v1",
		"gooo.metric.language.delivery-self-minted.guardrail.v1",
	}
}

func buildIndicators(summary Summary) []Indicator {
	coordinates := summary.Coordinates
	return []Indicator{
		indicator(MetricIDs()[0], ClassOutcome, ProofFoundation, "count-evidence-backed-delivery", coordinates.Satisfied, coordinates.Total),
		indicator(MetricIDs()[1], ClassOutcome, ProofCoherence, MetaOperation, coordinates.BasisPoints, 10000),
		indicator(MetricIDs()[2], ClassDriver, ProofFoundation, "bind-delivery-meta-operations", summary.MetaBindings, summary.MetaBindingsTotal),
		indicator(MetricIDs()[3], ClassDriver, ProofFoundation, "bind-source-artifact-manifest", summary.SourceReceipts, summary.SourceReceiptsTotal),
		indicator(MetricIDs()[4], ClassDriver, ProofCoherence, "project-reader-resolution-lattice", 3, 3),
		indicator(MetricIDs()[5], ClassGuardrail, ProofFoundation, "lower-unknown-delivery-evidence", coordinates.Unknown, 0),
		indicator(MetricIDs()[6], ClassGuardrail, ProofRegression, "deny-delivery-observer-effects", summary.Effects.RepositoryWrites, 0),
		indicator(MetricIDs()[7], ClassGuardrail, ProofRegression, "reject-self-minted-delivery-credit", summary.SelfMintedCredits, 0),
	}
}

func indicator(id string, class IndicatorClass, proof ProofChoice, operation string, value, target int) Indicator {
	satisfied := value == target
	return Indicator{MetricID: id, Class: class, ProofChoice: proof, MetaOperation: operation,
		Value: value, Target: target, Satisfied: satisfied}
}
