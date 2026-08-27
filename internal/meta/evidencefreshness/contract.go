package evidencefreshness

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/evidencefreshness/model"

func CanonicalContract() model.Contract {
	return model.Contract{
		Schema:     model.ContractSchema,
		Scope:      model.Scope,
		SourcePath: "examples/evidence-freshness/main.gooo",
		BaseContext: model.Context{
			Schema: model.ContextSchema,
			Tuple: model.EvidenceTuple{
				Recipe:      "recipe:gooo-evidence-freshness/v1",
				Environment: "environment:go1.27/linux-amd64/hermetic",
				Runner:      "runner:github-actions/ubuntu-24.04",
				Verifier:    "verifier:evidence-freshness-decider/v1",
			},
			CurrentEpoch:        20260827,
			EnvironmentBoundary: "environment:go1.27/linux-amd64/hermetic",
			Consumer:            model.ConsumerID,
		},
		Cases: []model.CaseDefinition{
			caseDefinition("fresh-exact", "", model.StateFresh, model.DecisionPass, model.ResolutionExact,
				model.StageVerifier, "accept-current-evidence", "TUPLE_EXACT_AND_BOUNDARY_CURRENT", "FOUNDATION", "preserve-exact-claim"),
			caseDefinition("subject-changed", "subject", model.StateStale, model.DecisionFailClosed, model.ResolutionInvariant,
				model.StageSubject, "compare-subject", "SUBJECT_CHANGED", "COHERENCE", "locate-subject-staleness"),
			caseDefinition("material-changed", "material", model.StateStale, model.DecisionFailClosed, model.ResolutionInvariant,
				model.StageMaterial, "compare-material", "MATERIAL_CHANGED", "COHERENCE", "locate-material-staleness"),
			caseDefinition("recipe-changed", "recipe", model.StateStale, model.DecisionFailClosed, model.ResolutionInvariant,
				model.StageRecipe, "compare-recipe", "RECIPE_CHANGED", "COHERENCE", "locate-recipe-staleness"),
			caseDefinition("environment-changed", "environment", model.StateStale, model.DecisionFailClosed, model.ResolutionInvariant,
				model.StageEnvironment, "compare-environment", "ENVIRONMENT_CHANGED", "COHERENCE", "locate-environment-staleness"),
			caseDefinition("runner-changed", "runner", model.StateStale, model.DecisionFailClosed, model.ResolutionInvariant,
				model.StageRunner, "compare-runner", "RUNNER_CHANGED", "COHERENCE", "locate-runner-staleness"),
			caseDefinition("verifier-changed", "verifier", model.StateStale, model.DecisionFailClosed, model.ResolutionInvariant,
				model.StageVerifier, "compare-verifier", "VERIFIER_CHANGED", "REGRESSION", "locate-verifier-staleness"),
			caseDefinition("temporal-boundary-expired", "temporal-expired", model.StateStale, model.DecisionFailClosed, model.ResolutionInvariant,
				model.StageVerifier, "check-validity-boundary", "TEMPORAL_BOUNDARY_EXPIRED", "FOUNDATION", "enforce-temporal-boundary"),
			caseDefinition("unknown-subject", "unknown-subject", model.StateUnknown, model.DecisionFailClosed, model.ResolutionLower,
				model.StageSubject, "read-subject", "SUBJECT_UNKNOWN", "REGRESSION", "preserve-unknown-subject"),
			caseDefinition("unknown-verifier", "unknown-verifier", model.StateUnknown, model.DecisionFailClosed, model.ResolutionLower,
				model.StageVerifier, "read-verifier", "VERIFIER_UNKNOWN", "REGRESSION", "preserve-unknown-verifier"),
		},
		Metrics: []model.MetricDefinition{
			metric("gooo.metric.evidence-freshness.cases.v1", "OUTCOME", "FOUNDATION", "evaluate-freshness-contract", model.CaseTotal, model.CaseTotal),
			metric("gooo.metric.evidence-freshness.fresh.v1", "OUTCOME", "FOUNDATION", "preserve-fresh-claim", 1, 1),
			metric("gooo.metric.evidence-freshness.stale.v1", "GUARDRAIL", "REGRESSION", "reject-stale-claim", 7, 7),
			metric("gooo.metric.evidence-freshness.unknown.v1", "GUARDRAIL", "REGRESSION", "preserve-unknown-resolution", 2, 2),
			metric("gooo.metric.evidence-freshness.coupling-axes.v1", "DRIVER", "COHERENCE", "enumerate-six-coupling-axes", model.AxisTotal, model.AxisTotal),
			metric("gooo.metric.evidence-freshness.stage-attribution.v1", "OUTCOME", "COHERENCE", "attribute-stale-stage", model.AxisTotal, model.AxisTotal),
			metric("gooo.metric.evidence-freshness.transitions.v1", "OUTCOME", "FOUNDATION", "preserve-claim-transition", model.TransitionTotal, model.TransitionTotal),
			metric("gooo.metric.evidence-freshness.temporal-boundary.v1", "GUARDRAIL", "FOUNDATION", "enforce-temporal-boundary", 1, 1),
			metric("gooo.metric.evidence-freshness.read-only.v1", "GUARDRAIL", "REGRESSION", "deny-repository-writes", 1, 1),
			metric("gooo.metric.evidence-freshness.independent-decider.v1", "GUARDRAIL", "FOUNDATION", "separate-decider-import-graph", 1, 1),
		},
		NotClaimed: []string{
			"cryptographic signature authenticity",
			"full compiler semantic correctness",
			"general cache eviction or recomputation",
			"wall-clock time or scheduler freshness",
			"external side effects or mutation authority",
		},
	}
}

func caseDefinition(id, mutation, state, decision, resolution, stage, step, reason, proof, operation string) model.CaseDefinition {
	return model.CaseDefinition{ID: id, Mutation: mutation, ExpectedState: state,
		ExpectedDecision: decision, ExpectedResolution: resolution, ExpectedStage: stage,
		ExpectedStep: step, ExpectedReason: reason, ProofChoice: proof, MetaOperation: operation}
}

func metric(id, class, proof, operation string, expected, denominator int) model.MetricDefinition {
	return model.MetricDefinition{MetricID: id, Class: class, Producer: model.ProducerID,
		Consumer: model.ConsumerID, ProofChoice: proof, MetaOperation: operation,
		ExpectedNumerator: expected, Denominator: denominator}
}
