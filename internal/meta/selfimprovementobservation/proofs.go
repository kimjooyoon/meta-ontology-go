package selfimprovementobservation

func observationProofs(indicators []Indicator, evidenceDigest string) []Proof {
	return []Proof{
		buildProof("FOUNDATION", "the source receipt and Gooo observation contract are fixed", "bind-read-only-observation-foundation", evidenceDigest, indicators),
		buildProof("COHERENCE", "minimal value witnesses agree across meta operations and reader views", "compare-observation-projections", evidenceDigest, indicators),
		buildProof("REGRESSION", "counterexamples, effects, authority, and replay remain bounded", "guard-read-only-observation", evidenceDigest, indicators),
	}
}

func buildProof(choice, claim, operation, digest string, indicators []Indicator) Proof {
	passed, found := true, false
	for _, indicator := range indicators {
		if indicator.ProofChoice != choice {
			continue
		}
		found = true
		passed = passed && indicator.Satisfied
	}
	return Proof{Choice: choice, Claim: claim, MetaOperation: operation, EvidenceDigest: digest, Passed: found && passed}
}

func observationNonClaims() []string {
	return []string{
		"business correctness", "value-level computation", "production readiness",
		"performance beyond this runner and fixed sample set", "general-purpose code generation",
		"improvement candidate quality", "automatic execution or adoption",
	}
}
