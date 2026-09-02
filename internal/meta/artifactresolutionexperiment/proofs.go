package artifactresolutionexperiment

func buildProofs(indicators []Indicator, evidence string) []Proof {
	return []Proof{
		proof("FOUNDATION", "two registered projections expose fixed information resolutions",
			"bind-resolution-foundation", indicators, evidence),
		proof("COHERENCE", "both projections preserve the same operation semantics",
			"compare-resolution-semantics", indicators, evidence),
		proof("REGRESSION", "replay, unknown projectors, and effects remain bounded",
			"guard-resolution-counterexamples", indicators, evidence),
	}
}

func proof(choice, claim, operation string, indicators []Indicator, evidence string) Proof {
	passed := true
	for _, indicator := range indicators {
		if indicator.ProofChoice == choice && !indicator.Satisfied {
			passed = false
		}
	}
	return Proof{Choice: choice, Claim: claim, MetaOperation: operation,
		EvidenceDigest: evidence, Passed: passed}
}
