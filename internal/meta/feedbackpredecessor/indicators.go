package feedbackpredecessor

func indicators(report Report) []Indicator {
	summary := report.Summary
	return []Indicator{
		indicator("gooo.metric.meta.predecessor-feedback-readiness.coverage-bps.v1",
			ClassOutcome, 10000, "basis_points", RelationGreaterOrEqual, ProofCoherence,
			"select-predecessor-feedback", "SelectPredecessorFeedback",
			boolBasisPoints(report.Decision == DecisionSelected)),
		indicator("gooo.metric.meta.predecessor-cycle-continuity.coverage-bps.v1",
			ClassOutcome, 10000, "basis_points", RelationGreaterOrEqual, ProofCoherence,
			"emit-predecessor-continuity", "EmitPredecessorContinuity",
			boolBasisPoints(Consumable(report))),
		indicator("gooo.metric.meta.predecessor-exact-head.coverage-bps.v1",
			ClassDriver, 10000, "basis_points", RelationGreaterOrEqual, ProofFoundation,
			"observe-predecessor-head", "ObservePredecessorHead",
			boolBasisPoints(summary.ExactHeadCandidates > 0)),
		indicator("gooo.metric.meta.predecessor-canonical-run.coverage-bps.v1",
			ClassDriver, 10000, "basis_points", RelationGreaterOrEqual, ProofCoherence,
			"bind-predecessor-run", "BindPredecessorRun",
			boolBasisPoints(summary.CanonicalCandidates > 0)),
		indicator("gooo.metric.meta.predecessor-successful-run.coverage-bps.v1",
			ClassDriver, 10000, "basis_points", RelationGreaterOrEqual, ProofRegression,
			"verify-predecessor-run", "VerifyPredecessorRun",
			boolBasisPoints(summary.SuccessfulCandidates > 0)),
		indicator("gooo.metric.meta.predecessor-receipt-bound.coverage-bps.v1",
			ClassDriver, 10000, "basis_points", RelationGreaterOrEqual, ProofRegression,
			"bind-predecessor-receipt", "BindPredecessorReceipt",
			boolBasisPoints(summary.ReceiptBoundCandidates > 0)),
		indicator("gooo.metric.meta.predecessor-ambiguity.guardrail.v1",
			ClassGuardrail, 0, "candidates", RelationLessOrEqual, ProofCoherence,
			"reject-ambiguous-predecessor", "RejectAmbiguousPredecessor",
			summary.AmbiguousCandidates),
		indicator("gooo.metric.meta.predecessor-observer-writes.guardrail.v1",
			ClassGuardrail, 0, "repository_writes", RelationLessOrEqual, ProofFoundation,
			"preserve-read-only-selection", "PreserveReadOnlySelection",
			summary.RepositoryWrites),
	}
}

func indicator(metric, class string, target int, unit, relation, proof,
	operation, activity string, value int) Indicator {
	satisfied := relation == RelationGreaterOrEqual && value >= target ||
		relation == RelationLessOrEqual && value <= target
	return Indicator{MetricID: metric, Class: class, Target: target, Unit: unit,
		Relation: relation, ProofChoice: proof, Producer: "feedbackpredecessor.Select",
		Consumer: "self-improvement-cycle", MetaOperation: operation, Activity: activity,
		Value: value, Satisfied: satisfied}
}

func boolBasisPoints(value bool) int {
	if value {
		return 10000
	}
	return 0
}
