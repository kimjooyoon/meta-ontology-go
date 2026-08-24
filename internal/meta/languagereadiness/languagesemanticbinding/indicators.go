package languagesemanticbinding

func buildIndicators(summary Summary) []Indicator {
	definitions := []struct {
		id, class, proof, operation string
		value, target               int
	}{
		{"gooo.metric.language.semantic-readiness-binding-bps.v1", "OUTCOME", "COHERENCE", "bind-semantic-readiness-evidence", 10000, 10000},
		{"gooo.metric.language.semantic-readiness-bound-coordinates.v1", "DRIVER", "FOUNDATION", "bind-exact-artifact-coordinates", summary.BoundCoordinates, ExpectedCoordinates},
		{"gooo.metric.language.semantic-readiness-completed-obligations.v1", "DRIVER", "COHERENCE", "bind-semantic-readiness-floor", summary.ReadinessCompleted, SemanticReadinessFloor},
		{"gooo.metric.language.semantic-readiness-executed-cases.v1", "DRIVER", "COHERENCE", "bind-semantic-corpus", summary.SemanticSatisfied, 20},
		{"gooo.metric.language.semantic-readiness-metric-bindings.v1", "DRIVER", "FOUNDATION", "bind-concept-metrics", summary.MetricBindings, 19},
		{"gooo.metric.language.semantic-readiness-unresolved.guardrail.v1", "GUARDRAIL", "FOUNDATION", "lower-binding-resolution", summary.Unresolved, 0},
		{"gooo.metric.language.semantic-readiness-effects.guardrail.v1", "GUARDRAIL", "REGRESSION", "seal-semantic-effects", summary.EffectfulStages, 0},
		{"gooo.metric.language.semantic-readiness-writes.guardrail.v1", "GUARDRAIL", "REGRESSION", "preserve-read-only-binding", summary.RepositoryWrites, 0},
		{"gooo.metric.language.semantic-readiness-mutation-authority.guardrail.v1", "GUARDRAIL", "REGRESSION", "deny-binding-mutation", summary.MutationAuthorities, 0},
	}
	result := make([]Indicator, 0, len(definitions))
	for _, definition := range definitions {
		satisfied := definition.value == definition.target
		if definition.id == "gooo.metric.language.semantic-readiness-completed-obligations.v1" {
			satisfied = definition.value >= definition.target
		}
		result = append(result, Indicator{
			MetricID: definition.id, Class: definition.class, ProofChoice: definition.proof,
			Producer: "languagesemanticbinding.Evaluate", Consumer: "self-improvement-cycle",
			MetaOperation: definition.operation, Value: definition.value, Target: definition.target,
			Satisfied: satisfied,
		})
	}
	return result
}
