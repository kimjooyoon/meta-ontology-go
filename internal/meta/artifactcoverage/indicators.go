package artifactcoverage

func CanonicalIndicators() []Indicator {
	return []Indicator{
		indicator("gooo.metric.meta.canonical-operation-artifact.coverage-bps.v1", ClassOutcome,
			10000, "basis_points", RelationGreaterOrEqual, ProofCoherence,
			"artifactcoverage.Evaluate", "self-improvement-cycle",
			"measure-operation-artifact-coverage", "MeasureOperationArtifactCoverage"),
		indicator("gooo.metric.meta.exact-head-artifact.coverage-bps.v1", ClassDriver,
			10000, "basis_points", RelationGreaterOrEqual, ProofFoundation,
			"artifactcoverage.Observe", "artifactcoverage.Evaluate",
			"bind-operation-artifact-foundation", "BindOperationArtifactFoundation"),
		indicator("gooo.metric.meta.digest-bound-artifact.coverage-bps.v1", ClassDriver,
			10000, "basis_points", RelationGreaterOrEqual, ProofCoherence,
			"artifactcoverage.Observe", "artifactcoverage.Evaluate",
			"resolve-canonical-operation-artifact", "ResolveCanonicalOperationArtifact"),
		indicator("gooo.metric.meta.replay-bound-artifact.coverage-bps.v1", ClassDriver,
			10000, "basis_points", RelationGreaterOrEqual, ProofRegression,
			"artifactcoverage.Replay", "artifactcoverage.Evaluate",
			"replay-operation-artifact-coverage", "ReplayOperationArtifactCoverage"),
		indicator("gooo.metric.meta.uncovered-artifact-operations.guardrail.v1", ClassGuardrail,
			0, "operations", RelationLessOrEqual, ProofCoherence,
			"artifactcoverage.Evaluate", "self-improvement-cycle",
			"select-uncovered-operation", "SelectUncoveredOperation"),
		indicator("gooo.metric.meta.ambiguous-artifact-bindings.guardrail.v1", ClassGuardrail,
			0, "bindings", RelationLessOrEqual, ProofCoherence,
			"artifactcoverage.Validate", "artifactcoverage.Evaluate",
			"resolve-canonical-operation-artifact", "ResolveCanonicalOperationArtifact"),
		indicator("gooo.metric.meta.artifact-observer-writes.guardrail.v1", ClassGuardrail,
			0, "repository_writes", RelationLessOrEqual, ProofFoundation,
			"artifactcoverage.Observe", "artifactcoverage.Evaluate",
			"preserve-read-only-artifact-observation", "PreserveReadOnlyArtifactObservation"),
	}
}

func indicator(id string, class IndicatorClass, target int, unit string, relation Relation,
	proof ProofChoice, producer, consumer, operation, activity string,
) Indicator {
	return Indicator{MetricID: id, Class: class, Target: target, Unit: unit, Relation: relation,
		ProofChoice: proof, Producer: producer, Consumer: consumer,
		MetaOperation: operation, Activity: activity}
}
