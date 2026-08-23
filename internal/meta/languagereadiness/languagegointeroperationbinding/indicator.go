package languagegointeroperationbinding

func indicators(summary Summary) []Indicator {
	return []Indicator{
		newIndicator("gooo.metric.language.go-interoperation-readiness-binding-bps.v1", "OUTCOME", "COHERENCE", "bind-go-interoperation-readiness", summary.BoundCoordinates*10000/12, 10000),
		newIndicator("gooo.metric.language.go-interoperation-readiness-bound-coordinates.v1", "DRIVER", "FOUNDATION", "bind-exact-go-interoperation-coordinates", summary.BoundCoordinates, 12),
		newIndicator("gooo.metric.language.go-interoperation-readiness-completed.v1", "DRIVER", "COHERENCE", "bind-readiness-count", summary.ReadinessCompleted, 17),
		newIndicator("gooo.metric.language.go-interoperation-readiness-cases.v1", "DRIVER", "COHERENCE", "bind-go-interoperation-corpus", summary.InteropSatisfied, 24),
		newIndicator("gooo.metric.language.go-interoperation-readiness-metrics.v1", "DRIVER", "FOUNDATION", "bind-go-interoperation-metrics", summary.MetricBindings, 18),
		newIndicator("gooo.metric.language.go-interoperation-readiness-unresolved.guardrail.v1", "GUARDRAIL", "FOUNDATION", "lower-binding-resolution", summary.Unresolved, 0),
		newIndicator("gooo.metric.language.go-interoperation-readiness-effects.guardrail.v1", "GUARDRAIL", "REGRESSION", "seal-go-interoperation-effects", summary.EffectfulStages, 0),
		newIndicator("gooo.metric.language.go-interoperation-readiness-writes.guardrail.v1", "GUARDRAIL", "REGRESSION", "preserve-read-only-binding", summary.RepositoryWrites, 0),
		newIndicator("gooo.metric.language.go-interoperation-readiness-mutation.guardrail.v1", "GUARDRAIL", "REGRESSION", "deny-binding-mutation", summary.MutationAuthorities, 0),
	}
}

func newIndicator(id, class, proof, operation string, value, target int) Indicator {
	return Indicator{MetricID: id, Class: class, ProofChoice: proof,
		MetaOperation: operation, Value: value, Target: target, Satisfied: value == target}
}
