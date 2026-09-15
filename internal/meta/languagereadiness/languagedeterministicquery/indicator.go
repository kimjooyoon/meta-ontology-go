package languagedeterministicquery

func indicators(summary Summary, resolution Resolution) []Indicator {
	return []Indicator{
		indicator("gooo.metric.language.deterministic-query-bps.v1", "OUTCOME", "COHERENCE", "measure-query-readiness", summary.ReadinessBPS, 10000, resolution),
		indicator("gooo.metric.language.deterministic-query-binding-plans.v1", "DRIVER", "FOUNDATION", "bind-query-plan-registry", summary.BindingPlans, 28, resolution),
		indicator("gooo.metric.language.deterministic-query-law-plans.v1", "DRIVER", "REGRESSION", "execute-query-laws", summary.LawPlans, 4, resolution),
		indicator("gooo.metric.language.deterministic-query-canonical-replays.v1", "DRIVER", "COHERENCE", "replay-canonical-query-receipts", summary.CanonicalReplays, 56, resolution),
		indicator("gooo.metric.language.deterministic-query-permutation-replays.v1", "DRIVER", "COHERENCE", "replay-insertion-permutation", summary.PermutationReplays, 28, resolution),
		indicator("gooo.metric.language.deterministic-query-concept-bindings.v1", "DRIVER", "FOUNDATION", "query-concept-binding", summary.ConceptBindings, 1, resolution),
		indicator("gooo.metric.language.deterministic-query-code-bindings.v1", "DRIVER", "FOUNDATION", "query-code-bindings", summary.CodeBindings, 6, resolution),
		indicator("gooo.metric.language.deterministic-query-metric-bindings.v1", "DRIVER", "COHERENCE", "query-metric-bindings", summary.MetricBindings, 18, resolution),
		indicator("gooo.metric.language.deterministic-query-use-case-bindings.v1", "DRIVER", "COHERENCE", "query-use-case-bindings", summary.UseCaseBindings, 3, resolution),
		indicator("gooo.metric.language.deterministic-query-not-satisfied.guardrail.v1", "GUARDRAIL", "REGRESSION", "reject-unsatisfied-query-cases", summary.NotSatisfied, 0, resolution),
		indicator("gooo.metric.language.deterministic-query-unresolved.guardrail.v1", "GUARDRAIL", "FOUNDATION", "lower-query-resolution", summary.Unresolved, 0, resolution),
		indicator("gooo.metric.language.deterministic-query-registry-drift.guardrail.v1", "GUARDRAIL", "FOUNDATION", "bind-versioned-query-registry", summary.RegistryDrift, 0, resolution),
		indicator("gooo.metric.language.deterministic-query-candidate-promotions.guardrail.v1", "GUARDRAIL", "REGRESSION", "reject-candidate-promotion", summary.CandidatePromotions, 0, resolution),
		indicator("gooo.metric.language.deterministic-query-unknown-acceptances.guardrail.v1", "GUARDRAIL", "REGRESSION", "fail-closed-query-unknowns", summary.UnknownAcceptances, 0, resolution),
		indicator("gooo.metric.language.deterministic-query-graph-mutations.guardrail.v1", "GUARDRAIL", "REGRESSION", "preserve-query-graph", summary.GraphMutations, 0, resolution),
		indicator("gooo.metric.language.deterministic-query-effectful-stages.guardrail.v1", "GUARDRAIL", "REGRESSION", "seal-query-effects", summary.EffectfulStages, 0, resolution),
		indicator("gooo.metric.language.deterministic-query-repository-writes.guardrail.v1", "GUARDRAIL", "REGRESSION", "preserve-read-only-query", 0, 0, resolution),
		indicator("gooo.metric.language.deterministic-query-mutation-authorities.guardrail.v1", "GUARDRAIL", "REGRESSION", "deny-query-mutation-authority", 0, 0, resolution),
	}
}

func indicator(metricID, class, proofChoice, operation string, value, target int, resolution Resolution) Indicator {
	return Indicator{
		MetricID: metricID, Class: class, ProofChoice: proofChoice,
		Producer: "languagedeterministicquery.Evaluate", Consumer: "self-improvement-cycle",
		MetaOperation: operation, Resolution: resolution,
		Value: value, Target: target, Satisfied: value == target,
	}
}
