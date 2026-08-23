package languagedeterministicquerybinding

func indicators(summary Summary) []Indicator {
	return []Indicator{
		newIndicator("gooo.metric.language.deterministic-query-readiness-binding-bps.v1", "OUTCOME", "COHERENCE", "bind-query-readiness", summary.BoundCoordinates*10000/FixedCoordinates, 10000),
		newIndicator("gooo.metric.language.deterministic-query-readiness-bound-coordinates.v1", "DRIVER", "FOUNDATION", "bind-exact-query-coordinates", summary.BoundCoordinates, 12),
		newIndicator("gooo.metric.language.deterministic-query-readiness-completed.v1", "DRIVER", "COHERENCE", "bind-readiness-count", summary.ReadinessCompleted, 16),
		newIndicator("gooo.metric.language.deterministic-query-readiness-cases.v1", "DRIVER", "COHERENCE", "bind-query-corpus", summary.QuerySatisfied, 32),
		newIndicator("gooo.metric.language.deterministic-query-readiness-metrics.v1", "DRIVER", "FOUNDATION", "bind-query-metrics", summary.MetricBindings, 18),
		newIndicator("gooo.metric.language.deterministic-query-readiness-unresolved.guardrail.v1", "GUARDRAIL", "FOUNDATION", "lower-binding-resolution", summary.Unresolved, 0),
		newIndicator("gooo.metric.language.deterministic-query-readiness-effects.guardrail.v1", "GUARDRAIL", "REGRESSION", "seal-query-effects", summary.EffectfulStages, 0),
		newIndicator("gooo.metric.language.deterministic-query-readiness-writes.guardrail.v1", "GUARDRAIL", "REGRESSION", "preserve-read-only-binding", summary.RepositoryWrites, 0),
		newIndicator("gooo.metric.language.deterministic-query-readiness-mutation.guardrail.v1", "GUARDRAIL", "REGRESSION", "deny-binding-mutation", summary.MutationAuthorities, 0),
	}
}

func newIndicator(metricID, class, proofChoice, operation string, value, target int) Indicator {
	return Indicator{
		MetricID: metricID, Class: class, ProofChoice: proofChoice,
		MetaOperation: operation, Value: value, Target: target, Satisfied: value == target,
	}
}
