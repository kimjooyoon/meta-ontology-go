package languagediagnosticprovenancebinding

func indicators(summary Summary) []Indicator {
	return []Indicator{
		newIndicator("gooo.metric.language.diagnostic-provenance-readiness-binding-bps.v1", "OUTCOME", "COHERENCE", "bind-diagnostic-provenance-readiness", summary.BoundCoordinates*10000/12, 10000),
		newIndicator("gooo.metric.language.diagnostic-provenance-readiness-bound-coordinates.v1", "DRIVER", "FOUNDATION", "bind-exact-diagnostic-coordinates", summary.BoundCoordinates, 12),
		newFloorIndicator("gooo.metric.language.diagnostic-provenance-readiness-completed.v1", "DRIVER", "COHERENCE", "bind-readiness-floor", summary.ReadinessCompleted, 18),
		newIndicator("gooo.metric.language.diagnostic-provenance-readiness-cases.v1", "DRIVER", "COHERENCE", "bind-diagnostic-corpus", summary.ProvenanceSatisfied, 18),
		newIndicator("gooo.metric.language.diagnostic-provenance-readiness-metrics.v1", "DRIVER", "FOUNDATION", "bind-diagnostic-metrics", summary.MetricBindings, 18),
		newIndicator("gooo.metric.language.diagnostic-provenance-readiness-unresolved.guardrail.v1", "GUARDRAIL", "FOUNDATION", "lower-binding-resolution", summary.Unresolved, 0),
		newIndicator("gooo.metric.language.diagnostic-provenance-readiness-effects.guardrail.v1", "GUARDRAIL", "REGRESSION", "seal-diagnostic-effects", summary.EffectfulStages, 0),
		newIndicator("gooo.metric.language.diagnostic-provenance-readiness-writes.guardrail.v1", "GUARDRAIL", "REGRESSION", "preserve-read-only-binding", summary.RepositoryWrites, 0),
		newIndicator("gooo.metric.language.diagnostic-provenance-readiness-mutation.guardrail.v1", "GUARDRAIL", "REGRESSION", "deny-binding-mutation", summary.MutationAuthorities, 0),
	}
}

func newIndicator(id, class, proof, operation string, value, target int) Indicator {
	return Indicator{
		MetricID: id, Class: class, ProofChoice: proof,
		MetaOperation: operation, Value: value, Target: target,
		Satisfied: value == target,
	}
}

func newFloorIndicator(id, class, proof, operation string, value, target int) Indicator {
	indicator := newIndicator(id, class, proof, operation, value, target)
	indicator.Satisfied = value >= target
	return indicator
}
