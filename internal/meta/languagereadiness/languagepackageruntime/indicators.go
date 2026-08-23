package languagepackageruntime

func indicators(summary Summary, resolution Resolution) []Indicator {
	return []Indicator{
		indicator("gooo.metric.language.package-runtime-readiness-bps.v1", "OUTCOME", "COHERENCE", "evaluate-package-runtime", summary.Satisfied*10000/FixedTotal, 10000, resolution),
		indicator("gooo.metric.language.package-runtime-positive-paths.v1", "OUTCOME", "COHERENCE", "execute-package-runtime", summary.PositivePaths, 10, resolution),
		indicator("gooo.metric.language.package-runtime-guardrail-rejections.v1", "OUTCOME", "REGRESSION", "reject-invalid-package-runtime", summary.GuardrailRejections, 8, resolution),
		indicator("gooo.metric.language.package-runtime-packages.v1", "DRIVER", "FOUNDATION", "compile-package-graph", summary.Packages, 40, resolution),
		indicator("gooo.metric.language.package-runtime-sources.v1", "DRIVER", "FOUNDATION", "lower-package-sources", summary.Sources, 50, resolution),
		indicator("gooo.metric.language.package-runtime-imports.v1", "DRIVER", "COHERENCE", "resolve-package-imports", summary.Imports, 40, resolution),
		indicator("gooo.metric.language.package-runtime-initializations.v1", "DRIVER", "COHERENCE", "order-package-initialization", summary.Initializations, 40, resolution),
		indicator("gooo.metric.language.package-runtime-entry-bindings.v1", "DRIVER", "COHERENCE", "resolve-entry-contract", summary.EntryBindings, 10, resolution),
		indicator("gooo.metric.language.package-runtime-semantic-bindings.v1", "DRIVER", "FOUNDATION", "bind-semantic-digests", summary.SemanticBindings, 50, resolution),
		indicator("gooo.metric.language.package-runtime-replays.v1", "DRIVER", "REGRESSION", "replay-runtime-image", summary.CanonicalReplays, 10, resolution),
		indicator("gooo.metric.language.package-runtime-order-invariants.v1", "DRIVER", "REGRESSION", "permute-runtime-input", summary.OrderInvariantReplays, 3, resolution),
		indicator("gooo.metric.language.package-runtime-unresolved.guardrail.v1", "GUARDRAIL", "FOUNDATION", "lower-runtime-resolution", summary.Unresolved, 0, resolution),
		indicator("gooo.metric.language.package-runtime-unknown.guardrail.v1", "GUARDRAIL", "FOUNDATION", "reject-unknown-runtime-observation", summary.UnknownObservations, 0, resolution),
		indicator("gooo.metric.language.package-runtime-invalid-acceptance.guardrail.v1", "GUARDRAIL", "REGRESSION", "reject-invalid-runtime", summary.InvalidAcceptances, 0, resolution),
		indicator("gooo.metric.language.package-runtime-graph-acceptance.guardrail.v1", "GUARDRAIL", "REGRESSION", "reject-invalid-package-graph", summary.GraphAcceptances, 0, resolution),
		indicator("gooo.metric.language.package-runtime-source-acceptance.guardrail.v1", "GUARDRAIL", "REGRESSION", "reject-invalid-package-source", summary.SourceAcceptances, 0, resolution),
		indicator("gooo.metric.language.package-runtime-entry-acceptance.guardrail.v1", "GUARDRAIL", "REGRESSION", "reject-invalid-entry", summary.EntryAcceptances, 0, resolution),
		indicator("gooo.metric.language.package-runtime-effects.guardrail.v1", "GUARDRAIL", "REGRESSION", "seal-runtime-effects", summary.EffectfulOperations+summary.RepositoryWrites+summary.MutationAuthorities, 0, resolution),
	}
}

func indicator(id, class, proof, operation string, value, target int, resolution Resolution) Indicator {
	return Indicator{MetricID: id, Class: class, ProofChoice: proof, MetaOperation: operation,
		Producer: "language-package-runtime-witness", Consumer: "language-readiness-witness",
		Value: value, Target: target, Resolution: resolution,
		Satisfied: resolution == ResolutionExact && value == target}
}
