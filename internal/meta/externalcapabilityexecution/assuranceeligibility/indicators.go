package assuranceeligibility

func buildIndicators(report Report) []Indicator {
	summary := report.Summary
	result := make([]Indicator, 0, 18)
	for _, artifact := range report.Artifacts {
		value := 0
		if artifact.Exact {
			value = 1
		}
		result = append(result, indicator("gooo.metric.evidence.external-eligibility."+artifact.Name+".v1",
			"DRIVER", "FOUNDATION", "consume-"+artifact.Name, "artifacts", "GREATER_OR_EQUAL",
			report.Resolution, value, 1))
	}
	result = append(result,
		indicator("gooo.metric.capability.external-eligibility.v1", "OUTCOME", "COHERENCE",
			MetaOperation, "paths", "GREATER_OR_EQUAL", report.Resolution, summary.EligiblePaths, 1),
		indicator("gooo.metric.assurance.external-projected-operating.v1", "OUTCOME", "COHERENCE",
			"project-assurance", "obligations", "GREATER_OR_EQUAL", report.Resolution, summary.ProjectedOperating, 12),
		indicator("gooo.metric.capability.external-selected-outcomes.v1", "OUTCOME", "COHERENCE",
			"execute-selected-capabilities", "outcomes", "GREATER_OR_EQUAL", report.Resolution,
			summary.CapabilityOutcomes, 3),
		indicator("gooo.metric.execution.external-selected-replay.v1", "OUTCOME", "COHERENCE",
			"replay-selected-capabilities", "executions", "GREATER_OR_EQUAL", report.Resolution,
			summary.ExternalExecutions, 4),
		indicator("gooo.metric.guardrail.external-parent-passes.v1", "GUARDRAIL", "REGRESSION",
			"preserve-parent-result", "checks", "GREATER_OR_EQUAL", report.Resolution, summary.ParentCompleted, 6),
		indicator("gooo.metric.guardrail.external-parent-failures.v1", "GUARDRAIL", "REGRESSION",
			"preserve-parent-failure", "checks", "GREATER_OR_EQUAL", report.Resolution, summary.ParentKnownFailures, 2),
		indicator("gooo.metric.guardrail.external-capability-suite.v1", "GUARDRAIL", "REGRESSION",
			"replay-capability-suite", "cases", "GREATER_OR_EQUAL", report.Resolution,
			summary.CapabilitySuitePassed, 15),
		indicator("gooo.metric.epistemic.external-eligibility-unknown.v1", "GUARDRAIL", "REGRESSION",
			"preserve-unknown", "paths", "LESS_OR_EQUAL", report.Resolution, summary.UnknownPaths, 0),
		indicator("gooo.metric.effects.external-eligibility-writes.v1", "GUARDRAIL", "REGRESSION",
			"deny-writes", "writes", "LESS_OR_EQUAL", report.Resolution,
			summary.RepositoryWrites+summary.ExternalRepositoryWrites, 0),
		indicator("gooo.metric.effects.external-eligibility-mutations.v1", "GUARDRAIL", "REGRESSION",
			"deny-official-mutation", "mutations", "LESS_OR_EQUAL", report.Resolution, summary.OfficialMutations, 0),
		indicator("gooo.metric.effects.external-eligibility-promotions.v1", "GUARDRAIL", "REGRESSION",
			"deny-promotion", "promotions", "LESS_OR_EQUAL", report.Resolution, summary.Promotions, 0))
	return result
}

func indicator(id, class, proof, operation, unit, relation, resolution string, value, target int) Indicator {
	satisfied := value <= target
	if relation == "GREATER_OR_EQUAL" {
		satisfied = value >= target
	}
	return Indicator{MetricID: id, Class: class, ProofChoice: proof,
		Producer: "assuranceeligibility.Evaluate", Consumer: "language-assurance-activation-gate",
		MetaOperation: operation, Unit: unit, Relation: relation, Resolution: resolution,
		Value: value, Target: target, Satisfied: satisfied}
}
