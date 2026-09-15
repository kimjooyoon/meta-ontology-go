package languagegointeroperation

func indicators(summary Summary, resolution Resolution) []Indicator {
	ambient := summary.ImportAcceptances + summary.EffectfulStages
	return []Indicator{
		indicator(fixedMetricBindings[0], "OUTCOME", "COHERENCE", "measure-go-interoperation", summary.ReadinessBPS, 10000, resolution),
		indicator(fixedMetricBindings[1], "OUTCOME", "COHERENCE", "accept-proven-go-boundaries", summary.PositiveAccepted, 16, resolution),
		indicator(fixedMetricBindings[2], "OUTCOME", "REGRESSION", "reject-invalid-go-boundaries", summary.GuardrailRejections, 8, resolution),
		indicator(fixedMetricBindings[3], "DRIVER", "COHERENCE", "project-semantic-ir", summary.GeneratorProjections, 8, resolution),
		indicator(fixedMetricBindings[4], "DRIVER", "COHERENCE", "check-go-1.27-boundaries", summary.Go127Boundaries, 8, resolution),
		indicator(fixedMetricBindings[5], "DRIVER", "COHERENCE", "replay-canonical-go", summary.CanonicalReplays, 16, resolution),
		indicator(fixedMetricBindings[6], "DRIVER", "COHERENCE", "replay-go-type-identity", summary.TypeIdentityReplays, 16, resolution),
		indicator(fixedMetricBindings[7], "DRIVER", "FOUNDATION", "observe-generic-methods", summary.GenericMethods, 5, resolution),
		indicator(fixedMetricBindings[8], "DRIVER", "FOUNDATION", "observe-materialized-aliases", summary.AliasNodes, 2, resolution),
		indicator(fixedMetricBindings[9], "DRIVER", "COHERENCE", "bind-generator-source-maps", summary.SourceMaps, 8, resolution),
		indicator(fixedMetricBindings[10], "DRIVER", "FOUNDATION", "reify-go-ast", summary.ASTReifications, 32, resolution),
		indicator(fixedMetricBindings[11], "GUARDRAIL", "REGRESSION", "reject-unsatisfied-cases", summary.NotSatisfied, 0, resolution),
		indicator(fixedMetricBindings[12], "GUARDRAIL", "FOUNDATION", "lower-unknown-resolution", summary.Unresolved, 0, resolution),
		indicator(fixedMetricBindings[13], "GUARDRAIL", "REGRESSION", "reject-invalid-acceptance", summary.InvalidAcceptances, 0, resolution),
		indicator(fixedMetricBindings[14], "GUARDRAIL", "REGRESSION", "fail-closed-unknown-payload", summary.UnknownAcceptances, 0, resolution),
		indicator(fixedMetricBindings[15], "GUARDRAIL", "REGRESSION", "deny-ambient-authority", ambient, 0, resolution),
		indicator(fixedMetricBindings[16], "GUARDRAIL", "REGRESSION", "preserve-read-only-evaluation", 0, 0, resolution),
		indicator(fixedMetricBindings[17], "GUARDRAIL", "REGRESSION", "deny-mutation-authority", 0, 0, resolution),
	}
}

func indicator(id, class, proof, operation string, value, target int, resolution Resolution) Indicator {
	return Indicator{MetricID: id, Class: class, ProofChoice: proof,
		Producer: "languagegointeroperation.Evaluate", Consumer: "self-improvement-cycle",
		MetaOperation: operation, Resolution: resolution, Value: value, Target: target,
		Satisfied: value == target}
}
