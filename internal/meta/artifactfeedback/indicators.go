package artifactfeedback

func CanonicalIndicators() []Indicator {
	return []Indicator{
		indicator("gooo.metric.meta.next-cycle-feedback.readiness-bps.v1", ClassOutcome, 10000,
			"basis_points", RelationGreaterOrEqual, ProofCoherence, "artifactfeedback.Evaluate", "self-improvement-cycle", "measure-next-cycle-feedback", "MeasureNextCycleFeedback"),
		indicator("gooo.metric.meta.feedback-exact-head.coverage-bps.v1", ClassDriver, 10000,
			"basis_points", RelationGreaterOrEqual, ProofFoundation, "artifactfeedback.Observe", "artifactfeedback.Evaluate", "observe-operation-artifact-feedback", "ObserveOperationArtifactFeedback"),
		indicator("gooo.metric.meta.feedback-cycle-bound.coverage-bps.v1", ClassDriver, 10000,
			"basis_points", RelationGreaterOrEqual, ProofCoherence, "artifactfeedback.Join", "artifactfeedback.Evaluate", "join-cycle-artifact-feedback", "JoinCycleArtifactFeedback"),
		indicator("gooo.metric.meta.feedback-replay-bound.coverage-bps.v1", ClassDriver, 10000,
			"basis_points", RelationGreaterOrEqual, ProofRegression, "artifactfeedback.Replay", "artifactfeedback.Evaluate", "replay-operation-artifact-feedback", "ReplayOperationArtifactFeedback"),
		indicator("gooo.metric.meta.feedback-stale-inputs.guardrail.v1", ClassGuardrail, 0,
			"inputs", RelationLessOrEqual, ProofFoundation, "artifactfeedback.Observe", "artifactfeedback.Evaluate", "observe-operation-artifact-feedback", "ObserveOperationArtifactFeedback"),
		indicator("gooo.metric.meta.feedback-ambiguous-next-operations.guardrail.v1", ClassGuardrail, 0,
			"operations", RelationLessOrEqual, ProofCoherence, "artifactfeedback.Select", "self-improvement-cycle", "select-next-meta-operation", "SelectNextMetaOperation"),
		indicator("gooo.metric.meta.feedback-observer-writes.guardrail.v1", ClassGuardrail, 0,
			"repository_writes", RelationLessOrEqual, ProofFoundation, "artifactfeedback.Observe", "artifactfeedback.Evaluate", "preserve-read-only-feedback", "PreserveReadOnlyFeedback"),
	}
}

func indicator(metric string, class IndicatorClass, target int, unit string, relation Relation,
	proof ProofChoice, producer, consumer, operation, activity string,
) Indicator {
	return Indicator{MetricID: metric, Class: class, Target: target, Unit: unit,
		Relation: relation, ProofChoice: proof, Producer: producer, Consumer: consumer,
		MetaOperation: operation, Activity: activity}
}

