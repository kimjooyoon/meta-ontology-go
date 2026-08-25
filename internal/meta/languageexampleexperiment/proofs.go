package languageexampleexperiment

func proofs(values []Indicator) []Proof {
	return []Proof{
		makeProof("FOUNDATION", "Gooo sources, emitter registry, samples, and non-claims are fixed",
			"bind-experiment-foundation", values, "value.artifact-digest-integrity", "compiler.source-files",
			"compiler.gooo-definition-bps", "compiler.emitter-registry", "resource.samples",
			"resource.valid-samples", "guardrail.non-claims"),
		makeProof("COHERENCE", "the emitted operation agrees with the independent golden",
			"compare-operation-projection", values, "value.primary-artifact", "value.golden-match"),
		makeProof("REGRESSION", "replay, resources, unknown emitters, and effects remain bounded",
			"guard-experiment-counterexamples", values, "value.deterministic-replay", "guardrail.wall",
			"guardrail.rss", "guardrail.binary", "counterexample.unknown-emitter", "guardrail.effects"),
	}
}

func makeProof(choice, claim, operation string, values []Indicator, ids ...string) Proof {
	passed := true
	for _, id := range ids {
		matched := false
		for _, value := range values {
			if value.ID == id {
				matched, passed = true, passed && value.Satisfied
			}
		}
		passed = passed && matched
	}
	return Proof{Choice: choice, Claim: claim, MetaOperation: operation, Passed: passed}
}
