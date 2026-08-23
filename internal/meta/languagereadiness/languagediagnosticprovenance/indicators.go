package languagediagnosticprovenance

func indicators(summary Summary, resolution Resolution) []Indicator {
	sealedEffects := summary.EffectfulStages
	return []Indicator{
		indicator(fixedMetricBindings[0], "OUTCOME", "COHERENCE", "measure-diagnostic-provenance", summary.ReadinessBPS, 10000, resolution),
		indicator(fixedMetricBindings[1], "OUTCOME", "COHERENCE", "trace-positive-diagnostics", summary.Traced, 10, resolution),
		indicator(fixedMetricBindings[2], "OUTCOME", "REGRESSION", "reject-provenance-unknowns", summary.GuardrailRejections, 8, resolution),
		indicator(fixedMetricBindings[3], "DRIVER", "FOUNDATION", "capture-physical-position", summary.PhysicalPositions, 10, resolution),
		indicator(fixedMetricBindings[4], "DRIVER", "COHERENCE", "resolve-logical-position", summary.LogicalPositions, 10, resolution),
		indicator(fixedMetricBindings[5], "DRIVER", "COHERENCE", "bind-semantic-identity", summary.SemanticBindings, 4, resolution),
		indicator(fixedMetricBindings[6], "DRIVER", "COHERENCE", "project-lsp-diagnostic", summary.LSPProjections, 10, resolution),
		indicator(fixedMetricBindings[7], "DRIVER", "COHERENCE", "replay-canonical-trace", summary.CanonicalReplays, 10, resolution),
		indicator(fixedMetricBindings[8], "DRIVER", "FOUNDATION", "sort-source-diagnostics", summary.OrderedDiagnostics, 6, resolution),
		indicator(fixedMetricBindings[9], "DRIVER", "COHERENCE", "observe-line-directive-remap", summary.LineDirectiveRemaps, 1, resolution),
		indicator(fixedMetricBindings[10], "DRIVER", "FOUNDATION", "classify-type-error-hardness", summary.TypeClassifications, 3, resolution),
		indicator(fixedMetricBindings[11], "GUARDRAIL", "REGRESSION", "reject-unsatisfied-cases", summary.NotSatisfied, 0, resolution),
		indicator(fixedMetricBindings[12], "GUARDRAIL", "FOUNDATION", "lower-unknown-resolution", summary.Unresolved, 0, resolution),
		indicator(fixedMetricBindings[13], "GUARDRAIL", "REGRESSION", "reject-unknown-provenance", summary.UnknownAcceptances, 0, resolution),
		indicator(fixedMetricBindings[14], "GUARDRAIL", "REGRESSION", "reject-missing-source-map", summary.MissingMapAccepts, 0, resolution),
		indicator(fixedMetricBindings[15], "GUARDRAIL", "REGRESSION", "reject-ambiguous-source-map", summary.AmbiguousAccepts, 0, resolution),
		indicator(fixedMetricBindings[16], "GUARDRAIL", "REGRESSION", "reject-invalid-coordinate", summary.InvalidAcceptances, 0, resolution),
		indicator(fixedMetricBindings[17], "GUARDRAIL", "REGRESSION", "preserve-zero-effect-observation", sealedEffects, 0, resolution),
	}
}

func indicator(id, class, proof, operation string, value, target int, resolution Resolution) Indicator {
	return Indicator{
		MetricID: id, Class: class, ProofChoice: proof,
		Producer: "languagediagnosticprovenance.Evaluate",
		Consumer: "self-improvement-cycle", MetaOperation: operation,
		Resolution: resolution, Value: value, Target: target,
		Satisfied: value == target,
	}
}
