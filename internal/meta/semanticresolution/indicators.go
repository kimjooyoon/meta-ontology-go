package semanticresolution

func CanonicalIndicators() []Indicator {
	return []Indicator{
		indicator("gooo.metric.meta.next-semantic-resolution.readiness-bps.v1", ClassOutcome, 10000, "basis_points", RelationGreaterOrEqual, ProofCoherence, "semanticresolution.ResolveSemanticConflict", "artifactfeedback.Evaluate", "select-coarse-recovery-operation", "SelectCoarseRecoveryOperation"),
		indicator("gooo.metric.meta.semantic-conflict-exact-head.coverage-bps.v1", ClassDriver, 10000, "basis_points", RelationGreaterOrEqual, ProofFoundation, "semanticresolution.ObserveSemanticConflict", "semanticresolution.ResolveSemanticConflict", "observe-semantic-conflict", "ObserveSemanticConflict"),
		indicator("gooo.metric.meta.semantic-resolution-monotone.coverage-bps.v1", ClassDriver, 10000, "basis_points", RelationGreaterOrEqual, ProofCoherence, "semanticresolution.LowerSemanticResolution", "semanticresolution.ResolveSemanticConflict", "lower-semantic-resolution", "LowerSemanticResolution"),
		indicator("gooo.metric.meta.semantic-resolution-replay.coverage-bps.v1", ClassDriver, 10000, "basis_points", RelationGreaterOrEqual, ProofRegression, "semanticresolution.ReplayResolutionTransition", "artifactfeedback.Evaluate", "replay-resolution-transition", "ReplayResolutionTransition"),
		indicator("gooo.metric.meta.semantic-resolution-descents.guardrail.v1", ClassGuardrail, MaxResolutionDescents, "descents", RelationLessOrEqual, ProofFoundation, "semanticresolution.ResolveSemanticConflict", "artifactfeedback.Evaluate", "preserve-resolution-descent-bound", "PreserveResolutionDescentBound"),
		indicator("gooo.metric.meta.semantic-resolution-unresolved.guardrail.v1", ClassGuardrail, 0, "conflicts", RelationLessOrEqual, ProofCoherence, "semanticresolution.ResolveSemanticConflict", "artifactfeedback.Evaluate", "select-coarse-recovery-operation", "SelectCoarseRecoveryOperation"),
		indicator("gooo.metric.meta.semantic-resolution-writes.guardrail.v1", ClassGuardrail, 0, "repository_writes", RelationLessOrEqual, ProofFoundation, "semanticresolution.ObserveSemanticConflict", "artifactfeedback.Evaluate", "preserve-read-only-resolution", "PreserveReadOnlyResolution"),
	}
}

func indicator(metric string, class IndicatorClass, target int, unit string, relation Relation, proof ProofChoice, producer, consumer, operation, activity string) Indicator {
	return Indicator{MetricID: metric, Class: class, Target: target, Unit: unit, Relation: relation, ProofChoice: proof, Producer: producer, Consumer: consumer, MetaOperation: operation, Activity: activity}
}
