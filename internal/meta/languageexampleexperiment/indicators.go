package languageexampleexperiment

func indicators(summary Summary, fixed Fixed) []Indicator {
	effectValue := int64(summary.Effects.RepositoryWrites)
	if summary.Effects.MutationAuthority {
		effectValue++
	}
	return []Indicator{
		metric("value.primary-artifact", "OUTCOME", "FOUNDATION", "count-primary-artifacts", int64(summary.Value.PrimaryArtifacts), int64(fixed.PrimaryArtifacts)),
		metric("value.golden-match", "OUTCOME", "COHERENCE", "compare-domain-golden", int64(summary.Value.GoldenMatches), 1),
		metric("value.deterministic-replay", "OUTCOME", "REGRESSION", "compare-artifact-digests", int64(summary.Value.DeterministicReplays), int64(fixed.DeterministicReplays)),
		metric("compiler.source-files", "DRIVER", "FOUNDATION", "count-bound-gooo-sources", int64(summary.Compiler.SourceFiles), int64(fixed.SourceFiles)),
		metric("compiler.gooo-definition-bps", "DRIVER", "FOUNDATION", "measure-definition-language-ratio", int64(summary.Compiler.GoooDefinitionBPS), 10000),
		metric("compiler.emitter-registry", "DRIVER", "FOUNDATION", "count-registered-emitters", int64(summary.Compiler.RegisteredEmitters), int64(fixed.RegisteredEmitters)),
		metric("resource.samples", "DRIVER", "FOUNDATION", "count-runner-samples", int64(summary.Resources.Samples), int64(fixed.ResourceSamples)),
		metric("guardrail.wall", "GUARDRAIL", "REGRESSION", "bound-observed-wall-time", int64(summary.Resources.WallViolations), 0),
		metric("guardrail.rss", "GUARDRAIL", "REGRESSION", "bound-observed-peak-rss", int64(summary.Resources.RSSViolations), 0),
		metric("guardrail.binary", "GUARDRAIL", "REGRESSION", "bound-observed-binary-size", int64(summary.Resources.BinaryViolations), 0),
		metric("counterexample.unknown-emitter", "GUARDRAIL", "REGRESSION", "reject-unknown-emitter", int64(summary.Counterexamples.UnknownEmitterRejections), int64(fixed.UnknownEmitterRejections)),
		metric("guardrail.effects", "GUARDRAIL", "REGRESSION", "deny-repository-effects", effectValue, 0),
		metric("guardrail.non-claims", "GUARDRAIL", "FOUNDATION", "preserve-experiment-non-claims", int64(summary.NotClaimed), int64(fixed.NotClaimed)),
	}
}

func metric(id, class, proof, operation string, value, target int64) Indicator {
	return Indicator{ID: id, Class: class, ProofChoice: proof, MetaOperation: operation,
		Value: value, Target: target, Satisfied: value == target}
}
