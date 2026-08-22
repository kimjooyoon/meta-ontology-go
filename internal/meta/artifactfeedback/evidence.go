package artifactfeedback

func evaluateKPIs(indicators []Indicator, summary Summary) []KPI {
	values := make(map[string]int, len(indicators))
	values["gooo.metric.meta.next-cycle-feedback.readiness-bps.v1"] = summary.ReadinessBasisPoints
	values["gooo.metric.meta.feedback-exact-head.coverage-bps.v1"] = summary.ExactHeadInputs * 5000
	values["gooo.metric.meta.feedback-cycle-bound.coverage-bps.v1"] = summary.BoundInputs * 5000
	values["gooo.metric.meta.feedback-replay-bound.coverage-bps.v1"] = summary.ReplayBoundInputs * 5000
	values["gooo.metric.meta.feedback-stale-inputs.guardrail.v1"] = summary.StaleInputs
	values["gooo.metric.meta.feedback-ambiguous-next-operations.guardrail.v1"] = summary.AmbiguousNextOperations
	values["gooo.metric.meta.feedback-observer-writes.guardrail.v1"] = summary.RepositoryWrites
	result := make([]KPI, 0, len(indicators))
	for _, indicator := range indicators {
		value := values[indicator.MetricID]
		satisfied := value >= indicator.Target
		if indicator.Relation == RelationLessOrEqual {
			satisfied = value <= indicator.Target
		}
		result = append(result, KPI{Indicator: indicator, Value: value, Satisfied: satisfied})
	}
	return result
}

func feedbackProofs(summary Summary) []Proof {
	return []Proof{
		feedbackProof(ProofFoundation, "observe-operation-artifact-feedback", "ObserveOperationArtifactFeedback",
			summary.StaleInputs == 0 && summary.RepositoryWrites == 0, summary),
		feedbackProof(ProofCoherence, "join-cycle-artifact-feedback", "JoinCycleArtifactFeedback",
			summary.BoundInputs == summary.RequiredInputs && summary.AmbiguousNextOperations == 0, summary),
		feedbackProof(ProofRegression, "replay-operation-artifact-feedback", "ReplayOperationArtifactFeedback",
			summary.ReplayBoundInputs == summary.RequiredInputs, summary),
	}
}

func feedbackProof(choice ProofChoice, operation, activity string, satisfied bool, summary Summary) Proof {
	evidence := struct {
		Choice  ProofChoice `json:"choice"`
		Summary Summary     `json:"summary"`
	}{Choice: choice, Summary: summary}
	return Proof{Choice: choice, MetaOperation: operation, Activity: activity,
		Satisfied: satisfied, EvidenceDigest: digestJSON(evidence)}
}
