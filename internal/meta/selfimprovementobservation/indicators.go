package selfimprovementobservation

func observationIndicators(check validation, replay bool) []Indicator {
	return []Indicator{
		booleanIndicator("foundation.source-schema", "DRIVER", "FOUNDATION", "bind-language-report-schema", check.SourceSchema),
		booleanIndicator("foundation.exact-head", "DRIVER", "FOUNDATION", "bind-exact-language-report-head", check.ExactHead),
		booleanIndicator("foundation.source-digest", "DRIVER", "FOUNDATION", "verify-language-report-content-digest", check.SourceDigest),
		booleanIndicator("foundation.gooo-observation-contract", "DRIVER", "FOUNDATION", "compile-read-only-observation-activity", check.Contract),
		booleanIndicator("foundation.fixed-denominators", "DRIVER", "FOUNDATION", "bind-fixed-observation-denominators", check.FixedDenominators),
		booleanIndicator("coherence.minimal-value-state", "OUTCOME", "COHERENCE", "classify-minimal-value-receipt", check.MinimalValueState),
		booleanIndicator("coherence.value-witnesses", "OUTCOME", "COHERENCE", "bind-generated-value-witnesses", check.ValueWitnesses),
		booleanIndicator("coherence.compiler-witnesses", "DRIVER", "COHERENCE", "bind-gooo-definition-witnesses", check.CompilerWitnesses),
		booleanIndicator("coherence.resource-witnesses", "OUTCOME", "COHERENCE", "bind-runner-resource-witnesses", check.ResourceWitnesses),
		booleanIndicator("coherence.meta-operations", "DRIVER", "COHERENCE", "bind-indicators-to-meta-operations", check.MetaOperations),
		booleanIndicator("coherence.munchhausen-proofs", "DRIVER", "COHERENCE", "bind-foundation-coherence-regression", check.Proofs),
		booleanIndicator("coherence.audience-resolutions", "DRIVER", "COHERENCE", "bind-reader-dependent-resolutions", check.Views),
		booleanIndicator("regression.counterexample-coverage", "GUARDRAIL", "REGRESSION", "reject-six-semantic-counterexamples", check.Counterexamples),
		booleanIndicator("regression.no-source-effects", "GUARDRAIL", "REGRESSION", "deny-source-repository-effects", check.SourceEffects),
		booleanIndicator("regression.read-only-authority", "GUARDRAIL", "REGRESSION", "deny-candidate-execution-mutation-promotion", check.ReadOnlyAuthority),
		booleanIndicator("regression.canonical-replay", "GUARDRAIL", "REGRESSION", "replay-read-only-observation", replay),
	}
}

func booleanIndicator(id, class, choice, operation string, pass bool) Indicator {
	value := 0
	if pass {
		value = 1
	}
	return Indicator{ID: id, Class: class, ProofChoice: choice, MetaOperation: operation, Value: value, Target: 1, Satisfied: pass}
}
